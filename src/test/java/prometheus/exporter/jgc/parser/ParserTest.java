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

import com.microsoft.gctoolkit.io.SingleGCLogFile;
import java.nio.file.Path;
import java.util.List;
import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.MockitoAnnotations;
import prometheus.exporter.jgc.tool.Metrics;

public class ParserTest {
    @Mock private GCEventRecorder eventRecorder;

    @Before
    public void init() {
        MockitoAnnotations.openMocks(this);
    }

    @Test
    public void test1() throws Exception {
        GCLogAnalyser analyser =
                new GCLogAnalyser(new SingleGCLogFile(Path.of("src/test/resources/parser/g1.log")));
        analyser.loadAggregators(List.of(new GCEventAggregator(eventRecorder)));
        Assert.assertEquals(true, analyser.analyze());
        Mockito.verify(eventRecorder, Mockito.times(1)).collectG1GCEvent(Mockito.isNotNull());
    }

    @Test
    public void test2() throws Exception {
        GCLogAnalyser analyser =
                new GCLogAnalyser(
                        new SingleGCLogFile(Path.of("src/test/resources/parser/zgc.log")));
        analyser.loadAggregators(List.of(new GCEventAggregator(eventRecorder)));
        Assert.assertEquals(true, analyser.analyze());
        Mockito.verify(eventRecorder, Mockito.times(1)).collectZGCEvent(Mockito.isNotNull());
    }

    @Test
    public void test3() throws Exception {
        GCLogAnalyser analyser =
                new GCLogAnalyser(
                        new SingleGCLogFile(Path.of("src/test/resources/parser/cms.log")));
        analyser.loadAggregators(List.of(new GCEventAggregator(eventRecorder)));
        Assert.assertEquals(true, analyser.analyze());
        Mockito.verify(eventRecorder, Mockito.times(20)).collectGenerationalGCEvent(Mockito.isNotNull());
    }
}
