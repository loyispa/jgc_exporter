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

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.github.lalyos.jfiglet.FigletFont;
import io.prometheus.client.exporter.HTTPServer;
import java.io.File;
import java.io.IOException;
import java.net.InetSocketAddress;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.collector.CleanableCollectorRegistry;
import prometheus.exporter.jgc.collector.GCCollector;
import prometheus.exporter.jgc.collector.SystemCollector;
import prometheus.exporter.jgc.tailer.TailerManager;

public class Bootstrap {
    private static final Logger LOG = LoggerFactory.getLogger(Bootstrap.class);
    private final HTTPServer httpServer;
    private final TailerManager tailerManager;
    private final SystemCollector systemCollector;
    private final GCCollector gcCollector;

    public Bootstrap(Config config) throws IOException {
        String hostPort = config.getHostPort();
        String host = hostPort.split(":")[0];
        int port = Integer.parseInt(hostPort.split(":")[1]);
        this.httpServer =
                new HTTPServer(
                        new InetSocketAddress(host, port),
                        CleanableCollectorRegistry.DEFAULT,
                        false);
        this.systemCollector = new SystemCollector().register(CleanableCollectorRegistry.DEFAULT);
        this.gcCollector = new GCCollector().register(CleanableCollectorRegistry.DEFAULT);
        this.tailerManager = new TailerManager(config, gcCollector);
    }

    public static void main(String[] args) throws Exception {

        printBanner();

        Config config = loadConfig(args);

        new Bootstrap(config);
    }

    static void printBanner() throws IOException {
        StringBuilder banner = new StringBuilder();
        banner.append(FigletFont.convertOneLine("jgc_exporter"));
        String v = Bootstrap.class.getPackage().getImplementationVersion();
        String version = String.format("%55s", v == null ? "unknown" : v);
        banner.append(version);
        LOG.info("{}", banner);
    }

    static Config loadConfig(String[] args) throws IOException {

        if (args.length < 1) {
            throw new IllegalArgumentException("config.yaml");
        }

        String configPath = args[0];
        ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
        Config config = mapper.readValue(new File(configPath), Config.class);

        String hostPort = config.getHostPort();
        String host = hostPort.split(":")[0];
        int port = Integer.parseInt(hostPort.split(":")[1]);
        if (host == null || port < 1000 || port > 65535) {
            throw new IllegalArgumentException("hostPort");
        }

        if (config.getFileRegexPattern() == null && config.getFileGlobPattern() == null) {
            throw new IllegalArgumentException("must specify fileRegexPattern or fileGlobPattern");
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

        return config;
    }
}
