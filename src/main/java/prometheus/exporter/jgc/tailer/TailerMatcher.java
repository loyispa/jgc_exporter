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
import java.util.ArrayList;
import java.util.Collection;
import java.util.List;
import java.util.stream.Collectors;

public class TailerMatcher {
    private final List<TailerSource> sources = new ArrayList<>();

    public TailerMatcher(String regexPattern, String globPattern) {
        if (regexPattern != null) {
            String[] regexPatterns = regexPattern.split(",");
            for (String pattern : regexPatterns) {
                sources.add(new RegexTailerSource(pattern));
            }
        }

        if (globPattern != null) {
            String[] globPatterns = globPattern.split(",");
            for (String pattern : globPatterns) {
                sources.add(new GlobTailerSource(pattern));
            }
        }
    }

    public List<File> findMatchingFiles() {
        return sources.stream()
                .map(TailerSource::findMatchingFiles)
                .flatMap(Collection::stream)
                .collect(Collectors.toList());
    }

    @Override
    public String toString() {
        return "TailerMatcher{" + "sources=" + sources + '}';
    }
}
