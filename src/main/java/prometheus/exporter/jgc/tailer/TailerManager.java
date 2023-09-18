/*
 * Copyright (C) 2023 The  jgc_exporter Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package prometheus.exporter.jgc.tailer;

import com.google.common.util.concurrent.ThreadFactoryBuilder;
import java.io.File;
import java.io.IOException;
import java.util.*;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.tool.Config;

public class TailerManager {
    private static final Logger LOG = LoggerFactory.getLogger(TailerManager.class);
    private final TailerMatcher tailerMatcher;
    private final Map<File, Tailer> tailers = new HashMap<>();
    private final ScheduledExecutorService checker;
    private final Thread runner;
    private final Lock lock;
    private final Listener listener;
    private final int batchSize;
    private final int bufferSize;
    private final long idleTimeout;

    public TailerManager(Config config, Listener listener) {
        this.tailerMatcher = new TailerMatcher(config.getFileRegexPattern());
        this.idleTimeout = config.getIdleTimeout();
        this.batchSize = config.getBatchSize();
        this.bufferSize = config.getBufferSize();
        this.listener = Objects.requireNonNull(listener);
        this.lock = new ReentrantLock();
        this.checker =
                Executors.newSingleThreadScheduledExecutor(
                        new ThreadFactoryBuilder().setNameFormat("tail-checker").build());
        this.checker.scheduleAtFixedRate(new CheckerRunnable(), 0, 5, TimeUnit.SECONDS);
        this.runner = new Thread(new TailerRunnable(), "tail-runner");
        this.runner.setDaemon(true);
        this.runner.start();
    }

    private boolean isIdle(long lastModified) {
        return lastModified + idleTimeout < System.currentTimeMillis();
    }

    private class CheckerRunnable implements Runnable {
        @Override
        public void run() {
            lock.lock();
            try {
                final List<File> matchingFiles =
                        tailerMatcher.findMatchingFiles(f -> !isIdle(f.lastModified()));
                for (File file : matchingFiles) {
                    tailers.computeIfAbsent(
                            file,
                            f -> {
                                Tailer tailer = new Tailer(f, true, batchSize, bufferSize);
                                try {
                                    listener.onOpen(f);
                                } catch (Throwable t) {
                                    LOG.error("Open tailer fail: {}", f, t);
                                }
                                return tailer;
                            });
                }

                final Iterator<Tailer> iterator = tailers.values().iterator();
                while (iterator.hasNext()) {
                    Tailer tailer = iterator.next();
                    if (tailer.rotate()) {
                        LOG.info("Rotate {}", tailer);
                    } else if (isIdle(tailer.getLastUpdated()) && !tailer.needTail()) {
                        try {
                            tailer.close();
                        } catch (IOException e) {
                        } finally {
                            LOG.info("Remove tailer {}", tailer);
                            iterator.remove();
                            try {
                                listener.onClose(tailer.getFile());
                            } catch (Throwable t) {
                                LOG.error("Close tailer fail: {}", tailer.getFile(), t);
                            }
                        }
                    }
                }

            } catch (Throwable t) {
                LOG.error("Check tailer fail.", t);
            } finally {
                lock.unlock();
            }
        }
    }

    private class TailerRunnable implements Runnable {
        @Override
        public void run() {
            while (true) {
                int produceLines = 0;
                lock.lock();
                try {
                    for (Tailer tailer : tailers.values()) {
                        try {
                            File file = tailer.getFile();
                            List<String> lines = tailer.readLines();
                            for (String line : lines) {
                                listener.onRead(file, line);
                            }
                            produceLines += lines.size();
                        } catch (Throwable t) {
                            LOG.error("Read tailer fail: {}", tailer, t);
                        }
                    }
                } finally {
                    lock.unlock();
                }
                if (produceLines == 0) {
                    try {
                        TimeUnit.SECONDS.sleep(1);
                    } catch (InterruptedException ignore) {
                    }
                }
                LOG.debug("Produce {} lines", produceLines);
            }
        }
    }

    public interface Listener {
        void onOpen(File file);

        void onClose(File file);

        void onRead(File file, String line);
    }
}
