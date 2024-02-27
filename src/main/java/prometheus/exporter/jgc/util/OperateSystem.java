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
package prometheus.exporter.jgc.util;

import java.io.*;
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.nio.file.Files;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class OperateSystem {
    private static final Logger LOG = LoggerFactory.getLogger(OperateSystem.class);

    public static final String OS = System.getProperty("os.name").toLowerCase();

    private OperateSystem() {}

    public static boolean isUnixLike() {
        return isLinux() || isUnix() || isMacOS();
    }

    public static boolean isLinux() {
        return OS.contains("linux");
    }

    public static boolean isUnix() {
        return OS.contains("nix");
    }

    public static boolean isMacOS() {
        return OS.contains("mac");
    }

    public static boolean isWindows() {
        return OS.contains("windows");
    }

    public static String getLocalHostName() {
        try {
            return InetAddress.getLocalHost().getHostName();
        } catch (UnknownHostException e) {
        }
        return "localhost";
    }

    public static Object getFileKey(File file) throws IOException {
        return Files.getAttribute(file.toPath(), "fileKey");
    }
}
