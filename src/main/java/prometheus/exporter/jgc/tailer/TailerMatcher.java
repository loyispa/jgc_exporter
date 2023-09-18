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

import com.google.common.collect.Lists;
import java.io.File;
import java.io.IOException;
import java.nio.file.*;
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.Objects;
import java.util.function.Predicate;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TailerMatcher {
    private static final Logger LOG = LoggerFactory.getLogger(TailerMatcher.class);
    private final List<TailerSource> sources = new ArrayList<>();

    public TailerMatcher(String filePattern) {
        String[] filePatterns = Objects.requireNonNull(filePattern).split(",");
        for (String pattern : filePatterns) {
            sources.add(new TailerSource(pattern));
        }
    }

    public List<File> findMatchingFiles(Predicate<File> predicate) {
        return sources.stream()
                .map(TailerSource::findMatchingFiles)
                .flatMap(Collection::stream)
                .filter(predicate)
                .collect(Collectors.toList());
    }

    static class TailerSource {
        private final String filePattern;
        private final File parentDir;
        private final DirectoryStream.Filter<Path> fileFilter;

        TailerSource(String filePattern) {
            this.filePattern = filePattern;
            File f = new File(filePattern);
            this.parentDir = f.getParentFile();
            String regex = f.getName();
            final PathMatcher matcher = FileSystems.getDefault().getPathMatcher("regex:" + regex);
            this.fileFilter =
                    entry -> matcher.matches(entry.getFileName()) && !Files.isDirectory(entry);
        }

        List<File> findMatchingFiles() {
            List<File> result = Lists.newArrayList();
            if (!parentDir.exists()) {
                return result;
            }
            try (DirectoryStream<Path> stream =
                    Files.newDirectoryStream(parentDir.toPath(), fileFilter)) {
                for (Path entry : stream) {
                    File file = entry.toFile();
                    result.add(file);
                }
            } catch (IOException e) {
                LOG.error("Find matching files fail: {} ", parentDir.toPath(), e);
            }
            LOG.debug("Find matching files: {}", result);
            return result;
        }

        @Override
        public String toString() {
            return "TailerSource{" + "filePattern='" + filePattern + '\'' + '}';
        }
    }

    @Override
    public String toString() {
        return "TailerMatcher{" + "sources=" + sources + '}';
    }
}
