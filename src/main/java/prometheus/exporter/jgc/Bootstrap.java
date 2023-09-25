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
package prometheus.exporter.jgc;

import static prometheus.exporter.jgc.tool.Metrics.*;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import io.prometheus.client.exporter.HTTPServer;
import java.io.File;
import java.io.IOException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.parser.ContinousGCLogFile;
import prometheus.exporter.jgc.tailer.TailerManager;
import prometheus.exporter.jgc.tool.Config;

public class Bootstrap {
    private static final Logger LOG = LoggerFactory.getLogger(Bootstrap.class);
    private final Config config;
    private final Map<File, ContinousGCLogFile> registry;
    private final HTTPServer httpServer;
    private final TailerManager tailerManager;

    public Bootstrap(Config config) throws IOException {
        this.config = config;
        this.registry = new ConcurrentHashMap<>();
        String hostPort = config.getHostPort();
        String host = hostPort.split(":")[0];
        int port = Integer.parseInt(hostPort.split(":")[1]);
        this.httpServer = new HTTPServer(host, port);
        this.tailerManager = new TailerManager(config, new LogFileListener());
    }

    public void run() throws InterruptedException {
        boolean needSleep = false;
        try {
            int analyzed =
                    registry.values().stream().mapToInt(file -> file.analyze() ? 1 : 0).sum();
            needSleep = analyzed == 0;
        } catch (Exception ex) {
            needSleep = true;
            LOG.error("Collect fail.", ex);
        } finally {
            if (needSleep) {
                TimeUnit.SECONDS.sleep(1);
            }
        }
    }

    private class LogFileListener implements TailerManager.Listener {
        @Override
        public void onOpen(File file) {
            LOG.info("Register file: {}", file);
            registry.computeIfAbsent(file, f -> new ContinousGCLogFile(file, config));
        }

        @Override
        public void onClose(File file) {
            LOG.info("Unregister file: {}", file);
            registry.remove(file);
        }

        @Override
        public void onRead(File file, String line) {
            GC_LOG_LINE.labels(file.getPath()).inc();
            registry.computeIfPresent(
                    file,
                    (f, log) -> {
                        log.append(line);
                        return log;
                    });
        }
    }

    public static void main(String[] args) throws Exception {

        if (args.length < 1) {
            throw new IllegalArgumentException("config.yaml");
        }

        ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
        Config config = mapper.readValue(new File(args[0]), Config.class);

        checkConfig(config);

        Bootstrap eventLoop = new Bootstrap(config);
        while (true) {
            eventLoop.run();
        }
    }

    static void checkConfig(Config config) {

        String hostPort = config.getHostPort();
        String host = hostPort.split(":")[0];
        int port = Integer.parseInt(hostPort.split(":")[1]);
        if (host == null || port < 1000 || port > 65535) {
            throw new IllegalArgumentException("port");
        }

        if (config.getFileRegexPattern() == null) {
            throw new IllegalArgumentException("fileRegexPattern");
        }

        if (config.getIdleTimeout() <= 0) {
            throw new IllegalArgumentException("idleTimeout");
        }

        if (config.getBufferSize() <= 0) {
            throw new IllegalArgumentException("bufferSize");
        }

        if (config.getBatchSize() <= 0) {
            throw new IllegalArgumentException("batchSize");
        }

        if (config.getAnalysePeriod() <= 0) {
            throw new IllegalArgumentException("analysePeriod");
        }

        if (config.getInflightRecordLength() <= 0) {
            throw new IllegalArgumentException("inflightRecordLength");
        }
    }
}
