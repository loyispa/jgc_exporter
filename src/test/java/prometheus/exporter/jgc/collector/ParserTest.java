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

import com.microsoft.gctoolkit.event.GarbageCollectionTypes;
import com.microsoft.gctoolkit.event.g1gc.G1Mixed;
import com.microsoft.gctoolkit.event.generational.GenerationalGCPauseEvent;
import com.microsoft.gctoolkit.event.zgc.ZGCCycle;
import com.microsoft.gctoolkit.message.ChannelName;
import java.io.File;
import org.junit.Before;
import org.junit.Test;
import org.mockito.Mockito;
import org.mockito.MockitoAnnotations;
import prometheus.exporter.jgc.collector.parser.GCAggregator;

public class ParserTest {

    @Before
    public void init() {
        MockitoAnnotations.openMocks(this);
    }

    @Test
    public void test1() throws Exception {
        File log = new File("src/test/resources/parser/g1.log");
        GCAggregator aggregator = Mockito.spy(new GCAggregator(log));
        G1Mixed event = Mockito.mock(G1Mixed.class);
        aggregator.publish(ChannelName.G1GC_PARSER_OUTBOX, event);
        Mockito.verify(aggregator, Mockito.times(1)).recordG1GCEvent(event);
    }

    @Test
    public void test2() throws Exception {
        File log = new File("src/test/resources/parser/zgc.log");
        GCAggregator aggregator = Mockito.spy(new GCAggregator(log));
        ZGCCycle event = Mockito.mock(ZGCCycle.class);
        aggregator.publish(ChannelName.ZGC_PARSER_OUTBOX, event);
        Mockito.verify(aggregator, Mockito.times(1)).recordZGCEvent(event);
    }

    @Test
    public void test3() throws Exception {
        File log = new File("src/test/resources/parser/cms.log");
        GCAggregator aggregator = Mockito.spy(new GCAggregator(log));
        GenerationalGCPauseEvent event = Mockito.mock(GenerationalGCPauseEvent.class);
        Mockito.when(event.getGarbageCollectionType())
                .thenReturn(GarbageCollectionTypes.CMSPausePhase);
        aggregator.publish(ChannelName.GENERATIONAL_HEAP_PARSER_OUTBOX, event);
        Mockito.verify(aggregator, Mockito.times(1)).recordGenerationalGCEvent(event);
    }
}
