/*
 * Copyright (C) 2024 The  jgc_exporter Authors
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

import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;
import com.google.common.util.concurrent.ThreadFactoryBuilder;
import java.io.File;
import java.util.*;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;
import java.util.function.Predicate;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.Config;
import prometheus.exporter.jgc.util.OperatingSystem;

public class TailerManager {
    private static final Logger LOG = LoggerFactory.getLogger(TailerManager.class);
    private final TailerMatcher tailerMatcher;
    private final Map<File, Tailer> registry = new HashMap<>();
    private final ScheduledExecutorService watcher;
    private final Lock lock;
    private final TailerListener listener;
    private final int batchSize;
    private final int bufferSize;
    private final long idleTimeout;
    private final int linesPerSecond;
    private final Predicate<Long> idleChecker;
    private final AtomicBoolean started;
    private final int readInterval;
    private final Cache<File, Long> invalidFiles;

    public TailerManager(Config config, TailerListener listener) {
        this.started = new AtomicBoolean(true);
        this.tailerMatcher =
                new TailerMatcher(config.getFileRegexPattern(), config.getFileGlobPattern());
        this.idleTimeout = config.getIdleTimeout();
        this.batchSize = config.getBatchSize();
        this.bufferSize = config.getBufferSize();
        this.linesPerSecond = config.getLinesPerSecond();
        this.readInterval = config.getReadInterval();
        this.listener = Objects.requireNonNull(listener);
        this.invalidFiles =
                CacheBuilder.newBuilder()
                        .expireAfterWrite(1, TimeUnit.HOURS)
                        .maximumSize(512)
                        .build();
        this.lock = new ReentrantLock();
        this.idleChecker = lastModified -> lastModified + idleTimeout < System.currentTimeMillis();
        this.watcher =
                Executors.newScheduledThreadPool(
                        2, new ThreadFactoryBuilder().setNameFormat("tail-watcher").build());
        this.watcher.scheduleAtFixedRate(
                new WatchRunnable(), 0, config.getWatchInterval(), TimeUnit.MILLISECONDS);
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
                        tailerMatcher.findMatchingFiles().stream()
                                .filter(f -> invalidFiles.getIfPresent(f) == null)
                                .filter(f -> idleChecker.negate().test(f.lastModified()))
                                .collect(Collectors.toList());
                for (File file : matchingFiles) {
                    try {
                        registry.computeIfAbsent(
                                file,
                                f -> {
                                    listener.onOpen(file);
                                    return newTailer(
                                            f, true, batchSize, bufferSize, linesPerSecond);
                                });
                    } catch (UnsupportedOperationException ignore) {
                        LOG.warn("Ignore unsupported file: {}", file);
                    } catch (Throwable t) {
                        LOG.error("Watch file error: {}", file, t);
                        invalidFiles.put(file, System.currentTimeMillis());
                    }
                }

                final Iterator<Tailer> iterator = registry.values().iterator();
                while (iterator.hasNext()) {
                    Tailer tailer = iterator.next();
                    if (idleChecker.test(tailer.lastModified())) {
                        try {
                            close(tailer);
                        } finally {
                            iterator.remove();
                        }
                    } else if (tailer.rotated()) {
                        try {
                            rotate(tailer);
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
        try {
            tailer.close();
        } finally {
            try {
                listener.onClose(tailer.getFile());
            } catch (Throwable t) {
                LOG.error("Close file failed: {}", tailer, t);
            }
        }
    }

    private void rotate(Tailer tailer) {
        try {
            tailer.close();
        } finally {
            try {
                listener.onRotate(tailer.getFile());
            } catch (Throwable t) {
                LOG.error("Rotate file failed: {}", tailer, t);
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
                LOG.debug("Read {} lines", produceLines);
                if (produceLines == 0) {
                    try {
                        TimeUnit.MILLISECONDS.sleep(readInterval);
                    } catch (InterruptedException ignore) {
                    }
                }
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
            watcher.shutdown();
        }
    }

    public static Tailer newTailer(
            File file, boolean seekToEnd, int batchSize, int bufferSize, int linesPerSecond) {
        if (OperatingSystem.isUnixLike()) {
            return new UnixLikeTailer(file, seekToEnd, batchSize, bufferSize, linesPerSecond);
        }
        if (OperatingSystem.isWindows()) {
            return new WindowsTailer(file, seekToEnd, batchSize, bufferSize, linesPerSecond);
        }
        LOG.warn("Unsupported OS: {}", OperatingSystem.OS);
        return new UnixLikeTailer(file, seekToEnd, batchSize, bufferSize, linesPerSecond);
    }
}
