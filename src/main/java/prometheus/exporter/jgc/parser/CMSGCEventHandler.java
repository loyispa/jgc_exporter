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

import static prometheus.exporter.jgc.metric.MetricRegistry.*;

import com.microsoft.gctoolkit.event.generational.*;
import com.microsoft.gctoolkit.event.jvm.JVMEvent;
import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.message.ChannelName;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.parser.*;
import java.io.File;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CMSGCEventHandler extends ParallelAndSerialGCEventHandler {
    private static final Logger LOG = LoggerFactory.getLogger(CMSGCEventHandler.class);

    public CMSGCEventHandler(File file, Diary diary) {
        super(file, diary);
    }

    @Override
    protected List<DataSourceParser> loadParsers() {
        if (diary.isUnifiedLogging()) {
            return Collections.singletonList(new UnifiedGenerationalParser());
        } else {
            return Arrays.asList(new CMSTenuredPoolParser(), new GenerationalHeapParser());
        }
    }

    @Override
    public void publish(ChannelName channel, JVMEvent event) {
        if (event instanceof CMSPhase
                || event instanceof ConcurrentModeFailure
                || event instanceof ConcurrentModeInterrupted) {

            if (shouldIgnore(channel, event)) {
                LOG.debug("Ignore CMS InitialMark or Remark due to GcToolkit bug");
                return;
            }

            recordCMSGCEvent((GenerationalGCEvent) event);
        } else {
            LOG.warn("{} published an unsupported event: {}", channel, event);
        }
    }

    private boolean shouldIgnore(ChannelName channel, JVMEvent event) {
        if (channel == ChannelName.CMS_TENURED_POOL_PARSER_OUTBOX) {
            if (event instanceof InitialMark || event instanceof CMSRemark) {
                return true;
            }
        }
        return false;
    }

    private String parseCategory(GenerationalGCEvent event) {
        if (event instanceof ConcurrentModeFailure) {
            return "CMSConcurrentModeFailure";
        } else if (event instanceof ConcurrentModeInterrupted) {
            return "CMSConcurrentModeInterrupted";
        } else if (event instanceof InitialMark) {
            return "CMSInitialMark";
        } else if (event instanceof CMSRemark) {
            return "CMSRemark";
        } else if (event instanceof AbortablePreClean) {
            return "CMSAbortablePreClean";
        } else if (event instanceof ConcurrentMark) {
            return "CMSConcurrentMark";
        } else if (event instanceof ConcurrentSweep) {
            return "CMSConcurrentSweep";
        } else if (event instanceof ConcurrentPreClean) {
            return "CMSConcurrentPreClean";
        } else if (event instanceof ConcurrentReset) {
            return "CMSConcurrentReset";
        }
        return "CMSUnknown";
    }

    private void recordCMSGCEvent(GenerationalGCEvent event) {

        String category = parseCategory(event);
        LOG.debug("Collect CMSGCEvent {}", category);
        GC_EVENT_DURATION.attach(this, path, host, category).observe(event.getDuration());
        GC_EVENT_LAST_MINUTE_DURATION.attach(this, path, host).observe(event.getDuration());

        if (event instanceof GenerationalGCPauseEvent) {
            GC_EVENT_PAUSE_DURATION.attach(this, path, host, category).observe(event.getDuration());
            recordCMSGCPauseEvent((GenerationalGCPauseEvent) event);
        }
    }

    private void recordCMSGCPauseEvent(GenerationalGCPauseEvent event) {

        recordGenerationalGCPauseEvent(event);

        if (event instanceof CMSPhase) {
            double classUnloadingProcessingTime = event.getClassUnloadingProcessingTime();
            double symbolTableProcessingTime = event.getSymbolTableProcessingTime();
            double stringTableProcessingTime = event.getStringTableProcessingTime();
            double symbolAndStringTableProcessingTime =
                    event.getSymbolAndStringTableProcessingTime();
            if (classUnloadingProcessingTime > 0.0) {
                CMS_CLASS_UNLOADING_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(classUnloadingProcessingTime);
            }
            if (symbolTableProcessingTime > 0.0) {
                CMS_SYMBOL_TABLE_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(symbolTableProcessingTime);
            }
            if (stringTableProcessingTime > 0.0) {
                CMS_STRING_TABLE_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(stringTableProcessingTime);
            }
            if (symbolAndStringTableProcessingTime > 0.0) {
                CMS_SYMBOL_AND_STRING_TABLE_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(symbolAndStringTableProcessingTime);
            }
        }
    }
}
