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
package prometheus.exporter.jgc.collector;

import static org.mockito.ArgumentMatchers.notNull;
import static org.mockito.Mockito.*;

import java.io.File;
import java.nio.file.Files;
import org.junit.Test;
import org.mockito.Mockito;
import prometheus.exporter.jgc.collector.parser.GCAggregator;

public class ParserTest {

    @Test
    public void test1() throws Exception {
        File log = new File("src/test/resources/parser/g1.log");
        GCAggregator aggregator =
                Mockito.mock(
                        GCAggregator.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(1)).recordG1GCEvent(notNull());
    }

    @Test
    public void test2() throws Exception {
        File log = new File("src/test/resources/parser/zgc.log");
        GCAggregator aggregator =
                Mockito.mock(
                        GCAggregator.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(1)).recordZGCEvent(notNull());
    }

    @Test
    public void test3() throws Exception {
        File log = new File("src/test/resources/parser/cms.log");
        GCAggregator aggregator =
                Mockito.mock(
                        GCAggregator.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(20)).recordGenerationalGCEvent(notNull());
    }
}
