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
package prometheus.exporter.jgc.tailer;

import java.io.*;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import org.junit.Assert;
import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.Config;

public class TailerTest {
    private static final Logger LOG = LoggerFactory.getLogger(TailerTest.class);

    @Test(timeout = 10000)
    public void testRead() throws Exception {

        File temp = File.createTempFile("test-read", "log");
        temp.deleteOnExit();
        PrintWriter pw = new PrintWriter(new OutputStreamWriter(new FileOutputStream(temp)));

        List<String> expectLines = new ArrayList<>();
        for (int i = 0; i < 100; ++i) {
            String line = "line:" + i;
            pw.println(line);
            expectLines.add(line);
        }
        pw.close();

        Tailer tailer =
                new Tailer(
                        temp,
                        false,
                        Config.DEFAULT_BATCH_SIZE,
                        Config.DEFAULT_BUFFER_SIZE,
                        Config.DEFAULT_LINES_PER_SECOND);

        List<String> actualLines = new ArrayList<>();

        while (true) {
            List<String> lines = tailer.readLines();
            if (lines.isEmpty()) {
                break;
            }
            actualLines.addAll(lines);
        }
        Assert.assertEquals(expectLines, actualLines);
    }

    @Test(timeout = 10000)
    public void testReadLimit() throws Exception {

        File temp = File.createTempFile("test-read", ".log");
        temp.deleteOnExit();
        PrintWriter pw = new PrintWriter(new OutputStreamWriter(new FileOutputStream(temp)));

        List<String> expectLines = new ArrayList<>();
        for (int i = 0; i < 2; ++i) {
            String line = "line:" + i;
            pw.println(line);
            expectLines.add(line);
        }
        pw.close();

        Tailer tailer =
                new Tailer(temp, false, Config.DEFAULT_BATCH_SIZE, Config.DEFAULT_BUFFER_SIZE, 1);

        int total = 1;
        while (total < 2) {
            List<String> lines = tailer.readLines();
            if (lines.isEmpty()) {
                continue;
            }
            total += lines.size();
            Assert.assertEquals(1, lines.size());
        }
        Assert.assertEquals(2, total);
    }

    @Test
    public void testRegexFind() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        List<File> expectFiles = new ArrayList<>();
        for (int i = 0; i < 10; ++i) {
            File temp = File.createTempFile("test-regex", ".log", tmpdir);
            expectFiles.add(temp);
            temp.deleteOnExit();
        }

        TailerMatcher matcher = new TailerMatcher(tmpdir.getPath() + "/test-regex.*\\.log", null);
        List<File> actualFiles = matcher.findMatchingFiles();
        expectFiles.sort(Comparator.comparing(File::getName));
        actualFiles.sort(Comparator.comparing(File::getName));
        Assert.assertEquals(expectFiles, actualFiles);
    }

    @Test
    public void testGlobFind() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        List<File> expectFiles = new ArrayList<>();
        for (int i = 0; i < 10; ++i) {
            File temp = File.createTempFile("test-glob", ".log", tmpdir);
            temp.deleteOnExit();
            expectFiles.add(temp);
        }

        TailerMatcher matcher = new TailerMatcher(null, tmpdir.getPath() + "/test-glob*.log");
        List<File> actualFiles = matcher.findMatchingFiles();
        expectFiles.sort(Comparator.comparing(File::getName));
        actualFiles.sort(Comparator.comparing(File::getName));
        Assert.assertEquals(expectFiles, actualFiles);
    }

    @Test(timeout = 600000)
    public void testListen() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        File temp = File.createTempFile("test-listen", ".log", tmpdir);
        temp.deleteOnExit();

        Config config = new Config();
        config.setFileRegexPattern(temp.getAbsolutePath());
        config.setWatchInterval(1000);

        CountDownLatch open = new CountDownLatch(1);
        CountDownLatch close = new CountDownLatch(1);
        TailerManager manager =
                new TailerManager(
                        config,
                        new TailerListener() {
                            @Override
                            public void onOpen(File file) {
                                open.countDown();
                            }

                            @Override
                            public void onClose(File file) {
                                close.countDown();
                            }

                            @Override
                            public void onRotate(File file) {}

                            @Override
                            public void onRead(File file, String line) {}
                        });

        open.await();
        manager.close();
        close.await();
    }

    @Test
    public void testChange() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        File file = File.createTempFile("test-rotate", ".log", tmpdir);
        Tailer tailer =
                new Tailer(
                        file,
                        true,
                        Config.DEFAULT_BATCH_SIZE,
                        Config.DEFAULT_BUFFER_SIZE,
                        Config.DEFAULT_LINES_PER_SECOND);

        Assert.assertEquals(tailer.rotated(), false);
        file.delete();
        file.createNewFile();
        Assert.assertEquals(tailer.rotated(), true);
    }

    @Test
    public void testTruncate() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        File file = File.createTempFile("test-rotate", ".log", tmpdir);
        Tailer tailer =
                new Tailer(
                        file,
                        true,
                        Config.DEFAULT_BATCH_SIZE,
                        Config.DEFAULT_BUFFER_SIZE,
                        Config.DEFAULT_LINES_PER_SECOND);

        Assert.assertEquals(tailer.rotated(), false);

        try (PrintWriter pw =
                new PrintWriter(new OutputStreamWriter(new FileOutputStream(file, false)))) {
            for (int i = 0; i < 10; ++i) {
                pw.println("line:" + i);
            }
        }

        while (!tailer.readLines().isEmpty())
            ;

        Assert.assertEquals(tailer.rotated(), false);

        try (PrintWriter pw =
                new PrintWriter(new OutputStreamWriter(new FileOutputStream(file, false)))) {
            for (int i = 0; i < 5; ++i) {
                pw.println("line:" + i);
            }
        }

        Assert.assertEquals(tailer.rotated(), true);
    }

    @Test(timeout = 60000)
    public void testIdle() throws Exception {

        File tmpdir = new File(System.getProperty("java.io.tmpdir"), "jgc");
        tmpdir.delete();
        tmpdir.mkdir();

        File temp = File.createTempFile("test-idle", ".log", tmpdir);
        temp.deleteOnExit();

        Config config = new Config();
        config.setFileGlobPattern(temp.getAbsolutePath());
        config.setIdleTimeout(3000);
        config.setWatchInterval(1000);

        CountDownLatch open = new CountDownLatch(1);
        CountDownLatch close = new CountDownLatch(1);
        TailerManager manager =
                new TailerManager(
                        config,
                        new TailerListener() {
                            @Override
                            public void onOpen(File file) {
                                open.countDown();
                            }

                            @Override
                            public void onClose(File file) {
                                close.countDown();
                            }

                            @Override
                            public void onRotate(File file) {}

                            @Override
                            public void onRead(File file, String line) {}
                        });

        open.await();
        Thread.sleep(5000);
        close.await();
        manager.close();
    }
}
