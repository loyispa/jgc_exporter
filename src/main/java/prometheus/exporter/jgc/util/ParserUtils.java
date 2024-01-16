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
package prometheus.exporter.jgc.util;

import static java.lang.Class.forName;

import com.microsoft.gctoolkit.io.SingleGCLogFile;
import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.message.DataSourceParser;
import java.io.File;
import java.lang.reflect.InvocationTargetException;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ParserUtils {
    private static final Logger LOG = LoggerFactory.getLogger(ParserUtils.class);

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
        "com.microsoft.gctoolkit.parser.ZGCParser"
    };

    private ParserUtils() {}

    public static List<DataSourceParser> findParsers(File file) {
        try {
            final SingleGCLogFile logFile = new SingleGCLogFile(file.toPath());

            boolean hasEnoughLogs = logFile.stream().skip(10).findAny().isPresent();
            if (!hasEnoughLogs) {
                return Collections.emptyList();
            }

            Diary diary = logFile.diary();
            if (diary.isG1GC()
                    || diary.isZGC()
                    || diary.isCMS()
                    || diary.isICMS()
                    || diary.isDefNew()
                    || diary.isSerialFull()
                    || diary.isPSOldGen()) {

                List<DataSourceParser> parsers =
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

                parsers.forEach(parser -> parser.diary(diary));
                return parsers;
            }
        } catch (Exception ex) {
            LOG.error("find parsers for {} error:", file, ex);
        }
        return Collections.emptyList();
    }
}
