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

import com.google.common.base.Splitter;
import com.google.common.collect.Lists;
import java.io.File;
import java.io.IOException;
import java.nio.file.*;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GlobTailerSource extends TailerSource {
    private static final Logger LOG = LoggerFactory.getLogger(TailerMatcher.class);
    private static final String globMetaChars = "\\*?[{";
    private final Path base;
    private final DirectoryStream.Filter<Path> fileFilter;

    public GlobTailerSource(String filePattern) {
        super(filePattern);
        this.base = getBase(filePattern);
        final PathMatcher matcher = FileSystems.getDefault().getPathMatcher("glob:" + filePattern);
        this.fileFilter = entry -> matcher.matches(entry) && !Files.isDirectory(entry);
    }

    @Override
    public List<File> findMatchingFiles() {
        List<File> result = Lists.newArrayList();

        try {
            Files.walkFileTree(
                    base,
                    new SimpleFileVisitor<>() {
                        @Override
                        public FileVisitResult visitFile(Path file, BasicFileAttributes attrs)
                                throws IOException {
                            if (fileFilter.accept(file)) {
                                result.add(file.toFile());
                            }
                            return super.visitFile(file, attrs);
                        }

                        @Override
                        public FileVisitResult visitFileFailed(Path file, IOException exc) {
                            return FileVisitResult.CONTINUE;
                        }
                    });
        } catch (IOException e) {
            LOG.error("Find matching files fail: {} ", filePattern, e);
        }

        LOG.debug("Find matching files: {}", result);
        return result;
    }

    private Path getBase(String pattern) {
        final String separator = FileSystems.getDefault().getSeparator();
        final List<String> dirs = Splitter.on(separator).splitToList(pattern);
        final StringBuilder path = new StringBuilder();

        for (int i = 0; i < dirs.size() - 1; ++i) {
            String dir = dirs.get(i);
            if (isWildcard(dir)) {
                break;
            }
            path.append(dir).append(separator);
        }

        return Path.of(path.toString()).toAbsolutePath();
    }

    private boolean isWildcard(String dir) {
        for (int j = 0; j < dir.length(); ++j) {
            char ch = dir.charAt(j);
            if (globMetaChars.indexOf(ch) != -1) {
                return true;
            }
        }
        return false;
    }
}
