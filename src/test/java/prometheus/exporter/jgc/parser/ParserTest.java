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
package prometheus.exporter.jgc.parser;

import static org.mockito.ArgumentMatchers.notNull;
import static org.mockito.Mockito.*;

import com.microsoft.gctoolkit.jvm.Diary;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import org.junit.Test;
import org.mockito.Mockito;

public class ParserTest {

    protected Diary getDiary(File file) throws IOException {
        return new GCEventHandlerMatcher(file).diary();
    }

    @Test
    public void testJdk8G1() throws Exception {
        File log = new File("src/test/resources/parser/jdk8-g1.log");
        Diary diary = getDiary(log);
        AbstractJVMEventHandler handler =
                Mockito.mock(
                        G1GCEventHandler.class,
                        withSettings()
                                .useConstructor(log, diary)
                                .defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(handler::consume);
        Mockito.verify(handler, Mockito.times(2)).publish(notNull(), notNull());
    }

    @Test
    public void testJdk11G1() throws Exception {
        File log = new File("src/test/resources/parser/jdk11-g1.log");
        Diary diary = getDiary(log);
        AbstractJVMEventHandler handler =
                Mockito.mock(
                        G1GCEventHandler.class,
                        withSettings()
                                .useConstructor(log, diary)
                                .defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(handler::consume);
        Mockito.verify(handler, Mockito.times(2)).publish(notNull(), notNull());
    }

    @Test
    public void testZGC() throws Exception {
        File log = new File("src/test/resources/parser/jdk11-zgc.log");
        Diary diary = getDiary(log);
        AbstractJVMEventHandler handler =
                Mockito.mock(
                        ZGCEventHandler.class,
                        withSettings()
                                .useConstructor(log, diary)
                                .defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(handler::consume);
        Mockito.verify(handler, Mockito.times(4)).publish(notNull(), notNull());
    }

    @Test
    public void testJdk8CMS() throws Exception {
        File log = new File("src/test/resources/parser/jdk8-cms-and-parnew.log");
        Diary diary = getDiary(log);
        AbstractJVMEventHandler handler =
                Mockito.mock(
                        CMSGCEventHandler.class,
                        withSettings()
                                .useConstructor(log, diary)
                                .defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(handler::consume);
        Mockito.verify(handler, Mockito.atLeast(8)).publish(isNotNull(), isNotNull());
    }

    @Test
    public void testJdk11CMS() throws Exception {
        File log = new File("src/test/resources/parser/jdk11-cms-and-parnew.log");
        Diary diary = getDiary(log);
        AbstractJVMEventHandler handler =
                Mockito.mock(
                        CMSGCEventHandler.class,
                        withSettings()
                                .useConstructor(log, diary)
                                .defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(handler::consume);
        Mockito.verify(handler, Mockito.times(7)).publish(isNotNull(), isNotNull());
    }

    @Test
    public void testJdk8ParallelOld() throws Exception {
        File log = new File("src/test/resources/parser/jdk8-parallel-old.log");
        Diary diary = getDiary(log);
        AbstractJVMEventHandler handler =
                Mockito.mock(
                        ParallelAndSerialGCEventHandler.class,
                        withSettings()
                                .useConstructor(log, diary)
                                .defaultAnswer(CALLS_REAL_METHODS));
        Files.lines(log.toPath()).forEach(handler::consume);
        Mockito.verify(handler, Mockito.times(6)).publish(isNotNull(), isNotNull());
    }
}
