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
package prometheus.exporter.jgc.collector;

import static org.mockito.ArgumentMatchers.notNull;
import static org.mockito.Mockito.*;

import java.io.File;
import java.nio.file.Files;
import org.junit.Test;
import org.mockito.Mockito;

public class ParserTest {

    @Test
    public void testJdk8G1() throws Exception {
        File log = new File("src/test/resources/parser/jdk8-g1.log");
        GCCollector aggregator =
                Mockito.mock(
                        GCCollector.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(2)).recordG1GCEvent(notNull());
    }

    @Test
    public void testJdk11G1() throws Exception {
        File log = new File("src/test/resources/parser/jdk11-g1.log");
        GCCollector aggregator =
                Mockito.mock(
                        GCCollector.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(2)).recordG1GCEvent(notNull());
    }

    @Test
    public void testZGC() throws Exception {
        File log = new File("src/test/resources/parser/jdk11-zgc.log");
        GCCollector aggregator =
                Mockito.mock(
                        GCCollector.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(4)).recordZGCEvent(notNull());
    }

    @Test
    public void testJdk8CMS() throws Exception {
        File log = new File("src/test/resources/parser/jdk8-cms-and-parnew.log");
        GCCollector aggregator =
                Mockito.mock(
                        GCCollector.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(7)).recordGenerationalGCEvent(isNotNull());
    }

    @Test
    public void testJdk11CMS() throws Exception {
        File log = new File("src/test/resources/parser/jdk11-cms-and-parnew.log");
        GCCollector aggregator =
                Mockito.mock(
                        GCCollector.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(6)).recordGenerationalGCEvent(isNotNull());
    }

    @Test
    public void testJdk8ParallelOld() throws Exception {
        File log = new File("src/test/resources/parser/jdk8-parallel-old.log");
        GCCollector aggregator =
                Mockito.mock(
                        GCCollector.class,
                        withSettings().useConstructor(log).defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(aggregator::receive);
        Mockito.verify(aggregator, Mockito.times(6)).recordGenerationalGCEvent(isNotNull());
    }
}
