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

import com.microsoft.gctoolkit.event.*;
import com.microsoft.gctoolkit.event.g1gc.*;
import com.microsoft.gctoolkit.event.jvm.JVMEvent;
import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.message.ChannelName;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.parser.PreUnifiedG1GCParser;
import com.microsoft.gctoolkit.parser.UnifiedG1GCParser;
import java.io.File;
import java.util.Collections;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class G1GCEventHandler extends AbstractJVMEventHandler {
    private static final Logger LOG = LoggerFactory.getLogger(G1GCEventHandler.class);

    public G1GCEventHandler(File file, Diary diary) {
        super(file, diary);
    }

    @Override
    protected List<DataSourceParser> loadParsers() {
        if (diary.isUnifiedLogging()) {
            return Collections.singletonList(new UnifiedG1GCParser());
        } else {
            return Collections.singletonList(new PreUnifiedG1GCParser());
        }
    }

    @Override
    public void publish(ChannelName channel, JVMEvent event) {
        if (event instanceof G1GCEvent) {
            recordG1GCEvent((G1GCEvent) event);
        } else {
            LOG.warn("{} published an unsupported event: {}", channel, event);
        }
    }

    void recordG1GCEvent(G1GCEvent event) {
        final String category = parseCategory(event);
        LOG.debug("Collect G1GCEvent {} ", category);

        GC_EVENT_DURATION.attach(this, path, host, category).observe(event.getDuration());
        GC_EVENT_LAST_MINUTE_DURATION.attach(this, path, host).observe(event.getDuration());

        if (event instanceof G1GCPauseEvent) {
            GC_EVENT_PAUSE_DURATION.attach(this, path, host, category).observe(event.getDuration());
            recordG1GCPauseEvent((G1GCPauseEvent) event);
        }
    }

    private String parseCategory(G1GCEvent event) {
        if (event instanceof G1GCPauseEvent) {
            if (event instanceof G1Young) {
                if (event instanceof G1YoungInitialMark) {
                    return "G1InitialMark";
                } else if (event instanceof G1Mixed) {
                    return "G1MixedGC";
                }
                return "G1YoungGC";
            } else if (event instanceof G1Cleanup) {
                return "G1Cleanup";
            } else if (event instanceof G1Remark) {
                return "G1Remark";
            } else if (event instanceof G1FullGC) {
                if (event instanceof G1SystemGC) {
                    return "G1SystemGC";
                }
                return "G1FullGC";
            }
        } else if (event instanceof G1GCConcurrentEvent) {
            if (event instanceof ConcurrentScanRootRegion) {
                return "G1ConcurrentScanRootRegion";
            } else if (event instanceof G1ConcurrentRebuildRememberedSets) {
                return "G1ConcurrentRebuildRememberedSets";
            } else if (event instanceof G1ConcurrentMarkResetForOverflow) {
                return "G1ConcurrentMarkResetForOverflow";
            } else if (event instanceof ConcurrentCreateLiveData) {
                return "G1ConcurrentCreateLiveData";
            } else if (event instanceof ConcurrentCompleteCleanup) {
                return "G1ConcurrentCompleteCleanup";
            } else if (event instanceof ConcurrentCleanupForNextMark) {
                return "G1ConcurrentCleanupForNextMark";
            } else if (event instanceof G1ConcurrentUndoCycle) {
                return "G1ConcurrentUndoCycle";
            } else if (event instanceof G1ConcurrentMark) {
                return "G1ConcurrentMark";
            } else if (event instanceof G1ConcurrentCleanup) {
                return "G1ConcurrentCleanup";
            } else if (event instanceof G1ConcurrentStringDeduplication) {
                return "G1ConcurrentStringDeduplication";
            } else if (event instanceof ConcurrentClearClaimedMarks) {
                return "G1ConcurrentClearClaimedMarks";
            }
        }
        return "G1Unknown";
    }

    private void recordG1GCPauseEvent(G1GCPauseEvent event) {

        MemoryPoolSummary heapSummary = event.getHeap();
        if (heapSummary != null) {
            HEAP_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getOccupancyAfterCollection() * 1024);
            HEAP_SIZE_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getSizeAfterCollection() * 1024);
            HEAP_SIZE_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getSizeBeforeCollection() * 1024);
            HEAP_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getOccupancyBeforeCollection() * 1024);
        }

        long youngOccupancyBeforeCollection = 0;
        long youngSizeBeforeCollection = 0;
        long youngOccupancyAfterCollection = 0;
        long youngSizeAfterCollection = 0;

        MemoryPoolSummary eden = event.getEden();
        if (eden != null) {
            if (eden.getOccupancyAfterCollection() >= 0) {
                G1_EDEN_OCCUPANCY_AFTER_COLLECTION
                        .attach(this, path, host)
                        .set(eden.getOccupancyAfterCollection() * 1024);
                youngOccupancyAfterCollection += eden.getOccupancyAfterCollection() * 1024;
            }
            if (eden.getOccupancyBeforeCollection() >= 0) {
                G1_EDEN_OCCUPANCY_BEFORE_COLLECTION
                        .attach(this, path, host)
                        .set(eden.getOccupancyBeforeCollection() * 1024);
                youngOccupancyBeforeCollection += eden.getOccupancyBeforeCollection() * 1024;
            }
            if (eden.getSizeBeforeCollection() > 0) {
                G1_EDEN_SIZE_BEFORE_COLLECTION
                        .attach(this, path, host)
                        .set(eden.getSizeBeforeCollection() * 1024);
                youngSizeBeforeCollection += eden.getSizeBeforeCollection() * 1024;
            }
            if (eden.getSizeAfterCollection() > 0) {
                G1_EDEN_SIZE_AFTER_COLLECTION
                        .attach(this, path, host)
                        .set(eden.getSizeAfterCollection() * 1024);
                youngSizeAfterCollection += eden.getSizeAfterCollection() * 1024;
            }
        }

        SurvivorMemoryPoolSummary survivor = event.getSurvivor();
        if (survivor != null) {
            if (survivor.getOccupancyAfterCollection() >= 0) {
                G1_SURVIVOR_HEAP_OCCUPANCY_AFTER_COLLECTION
                        .attach(this, path, host)
                        .set(survivor.getOccupancyAfterCollection() * 1024);
                youngOccupancyAfterCollection += survivor.getOccupancyAfterCollection() * 1024;
            }
            if (survivor.getOccupancyBeforeCollection() >= 0) {
                G1_SURVIVOR_HEAP_OCCUPANCY_BEFORE_COLLECTION
                        .attach(this, path, host)
                        .set(survivor.getOccupancyBeforeCollection() * 1024);

                youngOccupancyBeforeCollection += survivor.getOccupancyBeforeCollection() * 1024;
            }
            if (survivor.getSize() > 0) {
                G1_SURVIVOR_SIZE.attach(this, path, host).set(survivor.getSize() * 1024);
                youngSizeBeforeCollection += survivor.getSize() * 1024;
                youngSizeAfterCollection += survivor.getSize() * 1024;
            }
        }

        // young generation
        if (youngOccupancyBeforeCollection >= 0) {
            YOUNG_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(youngOccupancyBeforeCollection);
        }
        if (youngOccupancyAfterCollection >= 0) {
            YOUNG_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(youngOccupancyAfterCollection);
        }
        if (youngSizeBeforeCollection >= 0) {
            YOUNG_SIZE_BEFORE_COLLECTION.attach(this, path, host).set(youngSizeBeforeCollection);
        }
        if (youngSizeAfterCollection >= 0) {
            YOUNG_SIZE_AFTER_COLLECTION.attach(this, path, host).set(youngSizeAfterCollection);
        }

        MemoryPoolSummary metaspace = event.getPermOrMetaspace();
        if (metaspace != null) {
            METASPACE_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(metaspace.getOccupancyAfterCollection() * 1024);
            METASPACE_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(metaspace.getOccupancyBeforeCollection() * 1024);
            METASPACE_SIZE_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(metaspace.getSizeBeforeCollection() * 1024);
            METASPACE_SIZE_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(metaspace.getSizeAfterCollection() * 1024);
        }

        ReferenceGCSummary referenceGCSummary = event.getReferenceGCSummary();
        recordReferenceSummary(referenceGCSummary);

        RegionSummary edenRegion = event.getEdenRegionSummary();
        if (edenRegion != null) {
            if (edenRegion.getBefore() >= 0) {
                G1_EDEN_REGION_BEFORE.attach(this, path, host).set(edenRegion.getBefore());
            }
            if (edenRegion.getAfter() >= 0) {
                G1_EDEN_REGION_AFTER.attach(this, path, host).set(edenRegion.getAfter());
            }
            if (edenRegion.getAssigned() >= 0) {
                G1_EDEN_REGION_ASSIGN.attach(this, path, host).set(edenRegion.getAssigned());
            }
        }
        RegionSummary survivorRegion = event.getSurvivorRegionSummary();
        if (survivorRegion != null) {
            if (survivorRegion.getBefore() >= 0) {
                G1_SURVIVOR_REGION_BEFORE.attach(this, path, host).set(survivorRegion.getBefore());
            }
            if (survivorRegion.getAfter() >= 0) {
                G1_SURVIVOR_REGION_AFTER.attach(this, path, host).set(survivorRegion.getAfter());
            }
            if (survivorRegion.getAssigned() >= 0) {
                G1_SURVIVOR_REGION_ASSIGN
                        .attach(this, path, host)
                        .set(survivorRegion.getAssigned());
            }
        }
        RegionSummary oldRegion = event.getOldRegionSummary();
        if (oldRegion != null) {
            if (oldRegion.getBefore() >= 0) {
                G1_OLD_REGION_BEFORE.attach(this, path, host).set(oldRegion.getBefore());
            }
            if (oldRegion.getAfter() >= 0) {
                G1_OLD_REGION_AFTER.attach(this, path, host).set(oldRegion.getAfter());
            }
            if (oldRegion.getAssigned() >= 0) {
                G1_OLD_REGION_ASSIGN.attach(this, path, host).set(oldRegion.getAssigned());
            }
        }
        RegionSummary humongousRegion = event.getHumongousRegionSummary();
        if (humongousRegion != null) {
            if (humongousRegion.getBefore() >= 0) {
                G1_HUMONGOUS_REGION_BEFORE
                        .attach(this, path, host)
                        .set(humongousRegion.getBefore());
            }
            if (humongousRegion.getAfter() >= 0) {
                G1_HUMONGOUS_REGION_AFTER.attach(this, path, host).set(humongousRegion.getAfter());
            }
            if (humongousRegion.getAssigned() >= 0) {
                G1_HUMONGOUS_REGION_ASSIGN
                        .attach(this, path, host)
                        .set(humongousRegion.getAssigned());
            }
        }
        RegionSummary archiveRegion = event.getArchiveRegionSummary();
        if (archiveRegion != null) {
            if (archiveRegion.getBefore() >= 0) {
                G1_ARCHIVE_REGION_BEFORE.attach(this, path, host).set(archiveRegion.getBefore());
            }
            if (archiveRegion.getAfter() >= 0) {
                G1_ARCHIVE_REGION_AFTER.attach(this, path, host).set(archiveRegion.getAfter());
            }
            if (archiveRegion.getAssigned() >= 0) {
                G1_ARCHIVE_REGION_ASSIGN.attach(this, path, host).set(archiveRegion.getAssigned());
            }
        }
    }

    private void recordReferenceSummary(ReferenceGCSummary referenceGCSummary) {
        if (referenceGCSummary != null) {
            int softReferenceCount = referenceGCSummary.getSoftReferenceCount();
            SOFT_REFERENCE_COUNT.attach(this, path, host).set(softReferenceCount);
            double softReferencePauseTime = referenceGCSummary.getSoftReferencePauseTime();
            SOFT_REFERENCE_PAUSE_TIME.attach(this, path, host).observe(softReferencePauseTime);

            int weakReferenceCount = referenceGCSummary.getWeakReferenceCount();
            WEAK_REFERENCE_COUNT.attach(this, path, host).set(weakReferenceCount);
            double weakReferencePauseTime = referenceGCSummary.getWeakReferencePauseTime();
            WEAK_REFERENCE_PAUSE_TIME.attach(this, path, host).observe(weakReferencePauseTime);

            int finalReferenceCount = referenceGCSummary.getFinalReferenceCount();
            FINAL_REFERENCE_COUNT.attach(this, path, host).set(finalReferenceCount);
            double finalReferencePauseTime = referenceGCSummary.getFinalReferencePauseTime();
            FINAL_REFERENCE_PAUSE_TIME.attach(this, path, host).observe(finalReferencePauseTime);

            int phantomReferenceCount = referenceGCSummary.getPhantomReferenceCount();
            PHANTOM_REFERENCE_COUNT.attach(this, path, host).set(phantomReferenceCount);
            int phantomReferenceFreedCount = referenceGCSummary.getPhantomReferenceFreedCount();
            PHANTOM_REFERENCE_FREE_COUNT.attach(this, path, host).set(phantomReferenceFreedCount);
            double phantomReferencePauseTime = referenceGCSummary.getPhantomReferencePauseTime();
            PHANTOM_REFERENCE_PAUSE_TIME
                    .attach(this, path, host)
                    .observe(phantomReferencePauseTime);

            int jniWeakReferenceCount = referenceGCSummary.getJniWeakReferenceCount();
            JNI_WEAK_REFERENCE_COUNT.attach(this, path, host).set(jniWeakReferenceCount);
            double jniWeakReferencePauseTime = referenceGCSummary.getJniWeakReferencePauseTime();
            JNI_WEAK_REFERENCE_PAUSE_TIME
                    .attach(this, path, host)
                    .observe(jniWeakReferencePauseTime);
        }
    }
}
