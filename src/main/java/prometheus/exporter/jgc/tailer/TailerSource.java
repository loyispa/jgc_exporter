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
import java.util.List;

public abstract class TailerSource {
    protected final String filePattern;

    public TailerSource(String filePattern) {
        this.filePattern = filePattern;
    }

    public abstract List<File> findMatchingFiles();

    @Override
    public String toString() {
        return getClass().getSimpleName() + "{" + "filePattern='" + filePattern + '\'' + '}';
    }
}
