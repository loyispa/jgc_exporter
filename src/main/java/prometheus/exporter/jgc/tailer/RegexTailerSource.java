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

import com.google.common.collect.Lists;
import java.io.File;
import java.io.IOException;
import java.nio.file.*;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@Deprecated
public class RegexTailerSource extends TailerSource {
    private static final Logger LOG = LoggerFactory.getLogger(TailerMatcher.class);
    private final File base;
    private final DirectoryStream.Filter<Path> fileFilter;

    public RegexTailerSource(String filePattern) {
        super(filePattern);
        File f = new File(filePattern).getAbsoluteFile();
        this.base = f.getParentFile();
        String regex = f.getName();
        final PathMatcher matcher = FileSystems.getDefault().getPathMatcher("regex:" + regex);
        this.fileFilter =
                entry -> matcher.matches(entry.getFileName()) && !Files.isDirectory(entry);
    }

    @Override
    public List<File> findMatchingFiles() {
        List<File> result = Lists.newArrayList();
        if (!base.exists()) {
            return result;
        }
        try (DirectoryStream<Path> stream = Files.newDirectoryStream(base.toPath(), fileFilter)) {
            for (Path entry : stream) {
                File file = entry.toFile();
                result.add(file);
            }
        } catch (IOException e) {
            LOG.error("Find matching files fail: {} ", base.toPath(), e);
        }
        LOG.debug("Find matching files: {}", result);
        return result;
    }
}
