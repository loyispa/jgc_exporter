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
import java.util.*;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.tool.Config;

public class TailerManager {
    private static final Logger LOG = LoggerFactory.getLogger(TailerManager.class);
    private final TailerMatcher tailerMatcher;
    private final Map<File, Tailer> registry = new HashMap<>();
    private final ScheduledExecutorService watcher;
    private final Lock lock;
    private final Listener listener;
    private final int batchSize;
    private final int bufferSize;
    private final long idleTimeout;
    private final AtomicBoolean started = new AtomicBoolean(true);

    public TailerManager(Config config, Listener listener) {
        this.tailerMatcher = new TailerMatcher(config.getFileRegexPattern());
        this.idleTimeout = config.getIdleTimeout();
        this.batchSize = config.getBatchSize();
        this.bufferSize = config.getBufferSize();
        this.listener = Objects.requireNonNull(listener);
        this.lock = new ReentrantLock();
        this.watcher =
                Executors.newScheduledThreadPool(
                        2, new ThreadFactoryBuilder().setNameFormat("tail-watcher").build());
        this.watcher.scheduleAtFixedRate(new WatchRunnable(), 0, 5, TimeUnit.SECONDS);
        this.watcher.submit(new TailerRunnable());
    }

    private class WatchRunnable implements Runnable {
        @Override
        public void run() {
            if (!started.get()) {
                return;
            }
            lock.lock();
            try {
                final List<File> matchingFiles =
                        tailerMatcher.findMatchingFiles(
                                f -> f.lastModified() + idleTimeout > System.currentTimeMillis());
                for (File file : matchingFiles) {
                    registry.computeIfAbsent(
                            file,
                            f -> {
                                Tailer tailer =
                                        new Tailer(f, true, batchSize, bufferSize, idleTimeout);
                                try {
                                    listener.onOpen(f);
                                } catch (Throwable t) {
                                    LOG.error("Open file failed: {}", f, t);
                                }
                                return tailer;
                            });
                }

                final Iterator<Tailer> iterator = registry.values().iterator();
                while (iterator.hasNext()) {
                    Tailer tailer = iterator.next();
                    if (tailer.rotate()) {
                        LOG.info("Rotate file {}", tailer);
                    } else if (!tailer.needTail()) {
                        try {
                            close(tailer);
                        } finally {
                            iterator.remove();
                        }
                    }
                }

            } catch (Throwable t) {
                LOG.error("Watch file failed.", t);
            } finally {
                lock.unlock();
            }
        }
    }

    private void close(Tailer tailer) {
        File file = tailer.getFile();
        try {
            LOG.info("Remove file {}", file);
            tailer.close();
        } finally {
            try {
                listener.onClose(file);
            } catch (Throwable t) {
                LOG.error("Close file failed: {}", file, t);
            }
        }
    }

    private class TailerRunnable implements Runnable {
        @Override
        public void run() {
            while (started.get()) {
                int produceLines = 0;
                lock.lock();
                try {
                    for (Tailer tailer : registry.values()) {
                        try {
                            File file = tailer.getFile();
                            List<String> lines = tailer.readLines();
                            for (String line : lines) {
                                listener.onRead(file, line);
                            }
                            produceLines += lines.size();
                        } catch (Throwable t) {
                            LOG.error("Read file failed: {}", tailer, t);
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
                LOG.debug("Read {} lines", produceLines);
            }

            lock.lock();
            try {
                registry.values().forEach(TailerManager.this::close);
            } finally {
                registry.clear();
                lock.unlock();
            }
        }
    }

    public void close() {
        if (started.compareAndSet(true, false)) {
            LOG.info("TailerManager is closing.");
            watcher.shutdown();
            LOG.info("TailerManager is closed.");
        } else {
            LOG.warn("TailerManager is already closed.");
        }
    }

    public interface Listener {
        void onOpen(File file);

        void onClose(File file);

        void onRead(File file, String line);
    }
}
