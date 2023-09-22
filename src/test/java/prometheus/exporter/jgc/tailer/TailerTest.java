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
package prometheus.exporter.jgc.tailer;

import java.io.File;
import java.io.FileOutputStream;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.util.List;
import java.util.concurrent.Phaser;
import org.junit.Assert;
import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.tool.Config;

public class TailerTest {
    private static final Logger LOG = LoggerFactory.getLogger(TailerTest.class);

    @Test
    public void testRead() throws Exception {

        File temp = File.createTempFile("temp", "log");
        temp.deleteOnExit();
        PrintWriter pw = new PrintWriter(new OutputStreamWriter(new FileOutputStream(temp)));
        for (int i = 0; i < 100; ++i) {
            pw.println("line:" + i);
        }
        pw.close();

        Tailer tailer =
                new Tailer(
                        temp,
                        false,
                        Config.DEFAULT_BATCH_SIZE,
                        Config.DEFAULT_BUFFER_SIZE,
                        Config.DEFAULT_IDLE_TIMEOUT);
        int total = 0;
        while (true) {
            List<String> lines = tailer.readLines();
            total += lines.size();
            if (lines.isEmpty()) {
                break;
            }
        }
        Assert.assertEquals(total, 100);
    }

    @Test
    public void testRotate() throws Exception {

        File temp = File.createTempFile("temp", "log");

        Tailer tailer =
                new Tailer(
                        temp,
                        true,
                        Config.DEFAULT_BATCH_SIZE,
                        Config.DEFAULT_BUFFER_SIZE,
                        Config.DEFAULT_IDLE_TIMEOUT);

        Assert.assertEquals(tailer.rotate(), false);

        temp.delete();
        temp.createNewFile();

        Assert.assertEquals(tailer.rotate(), true);
    }

    @Test
    public void testFind() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        for (int i = 0; i < 10; ++i) {
            File temp = File.createTempFile("test", "log", tmpdir);
            temp.deleteOnExit();
        }

        TailerMatcher matcher = new TailerMatcher(tmpdir.getPath() + "/.*.log");
        int files = matcher.findMatchingFiles(f -> true).size();
        Assert.assertEquals(10, files);
    }

    @Test
    public void testListen() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        File temp = File.createTempFile("test", "log", tmpdir);
        temp.deleteOnExit();

        Config config = new Config();
        config.setFileRegexPattern(tmpdir.getPath() + "/.*.log");

        Phaser phaser = new Phaser(2);
        TailerManager manager =
                new TailerManager(
                        config,
                        new TailerManager.Listener() {
                            @Override
                            public void onOpen(File file) {
                                phaser.arriveAndDeregister();
                            }

                            @Override
                            public void onClose(File file) {
                                phaser.arriveAndDeregister();
                            }

                            @Override
                            public void onRead(File file, String line) {}
                        });

        manager.close();
        phaser.awaitAdvance(0);
    }
}
