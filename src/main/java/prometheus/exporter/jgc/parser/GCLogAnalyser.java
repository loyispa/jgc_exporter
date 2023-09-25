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
package prometheus.exporter.jgc.parser;

import static java.lang.Class.forName;

import com.google.common.annotations.VisibleForTesting;
import com.microsoft.gctoolkit.aggregator.Aggregation;
import com.microsoft.gctoolkit.aggregator.Aggregator;
import com.microsoft.gctoolkit.io.GCLogFile;
import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.jvm.JavaVirtualMachine;
import com.microsoft.gctoolkit.message.DataSourceChannel;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.message.JVMEventChannel;
import com.microsoft.gctoolkit.vertx.VertxDataSourceChannel;
import com.microsoft.gctoolkit.vertx.VertxJVMEventChannel;
import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.util.*;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GCLogAnalyser {
    private static final Logger LOG = LoggerFactory.getLogger(GCLogAnalyser.class);
    private static final String[] DEFAULT_PARSERS = {
        "com.microsoft.gctoolkit.parser.CMSTenuredPoolParser",
        "com.microsoft.gctoolkit.parser.GenerationalHeapParser",
        "com.microsoft.gctoolkit.parser.JVMEventParser",
        "com.microsoft.gctoolkit.parser.PreUnifiedG1GCParser",
        "com.microsoft.gctoolkit.parser.ShenandoahParser",
        "com.microsoft.gctoolkit.parser.SurvivorMemoryPoolParser",
        "com.microsoft.gctoolkit.parser.UnifiedG1GCParser",
        "com.microsoft.gctoolkit.parser.UnifiedGenerationalParser",
        "com.microsoft.gctoolkit.parser.UnifiedJVMEventParser",
        "com.microsoft.gctoolkit.parser.UnifiedSurvivorMemoryPoolParser",
        "prometheus.exporter.jgc.parser.zgc.ContinousZGCParser"
    };
    private final GCLogFile logFile;
    private JavaVirtualMachine javaVirtualMachine;
    private List<Aggregator<? extends Aggregation>> aggregators;
    private List<DataSourceParser> dataSourceParsers;

    public GCLogAnalyser(GCLogFile logFile) throws IOException {
        this.logFile = logFile;
        loadJavaVirtualMachine();
        loadAggregators();
        loadDataSourceParsers();
    }

    private void loadJavaVirtualMachine() {
        javaVirtualMachine = logFile.getJavaVirtualMachine();
    }

    private void loadAggregators() {
        this.aggregators =
                List.of(new GCEventAggregator(new GCEventRecorder(logFile.getPath().toString())));
    }

    @VisibleForTesting
    void loadAggregators(List<Aggregator<? extends Aggregation>> aggregators) {
        this.aggregators = Objects.requireNonNull(aggregators);
    }

    private void loadDataSourceParsers() throws IOException {
        Diary diary = logFile.diary();

        dataSourceParsers =
                ServiceLoader.load(DataSourceParser.class).stream()
                        .map(ServiceLoader.Provider::get)
                        .filter(dataSourceParser -> dataSourceParser.accepts(diary))
                        .collect(Collectors.toList());

        if (dataSourceParsers.isEmpty()) {
            dataSourceParsers =
                    Arrays.stream(DEFAULT_PARSERS)
                            .map(
                                    parserName -> {
                                        try {
                                            Class<?> clazz =
                                                    forName(
                                                            parserName,
                                                            true,
                                                            Thread.currentThread()
                                                                    .getContextClassLoader());
                                            return Optional.of(
                                                    clazz.getConstructors()[0].newInstance());
                                        } catch (ClassNotFoundException
                                                | InstantiationException
                                                | IllegalAccessException
                                                | InvocationTargetException e) {
                                            return Optional.empty();
                                        }
                                    })
                            .filter(Optional::isPresent)
                            .map(optional -> (DataSourceParser) optional.get())
                            .filter(dataSourceParser -> dataSourceParser.accepts(diary))
                            .collect(Collectors.toList());
        }

        for (DataSourceParser dataSourceParser : dataSourceParsers) {
            dataSourceParser.diary(diary);
        }
    }

    public boolean analyze() {
        try {
            DataSourceChannel dataSourceChannel = new VertxDataSourceChannel();
            JVMEventChannel jvmEventChannel = new VertxJVMEventChannel();
            for (DataSourceParser dataSourceParser : dataSourceParsers) {
                dataSourceChannel.registerListener(dataSourceParser);
                dataSourceParser.publishTo(jvmEventChannel);
            }
            javaVirtualMachine.analyze(aggregators, jvmEventChannel, dataSourceChannel);
            return true;
        } catch (Throwable t) {
            LOG.error("Analyze fail", t);
        }
        return false;
    }
}
