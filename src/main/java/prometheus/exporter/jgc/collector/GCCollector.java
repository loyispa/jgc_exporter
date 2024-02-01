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
package prometheus.exporter.jgc.collector;

import static prometheus.exporter.jgc.metric.MetricRegistry.*;

import com.google.common.annotations.VisibleForTesting;
import com.microsoft.gctoolkit.event.*;
import com.microsoft.gctoolkit.event.g1gc.*;
import com.microsoft.gctoolkit.event.generational.*;
import com.microsoft.gctoolkit.event.jvm.JVMEvent;
import com.microsoft.gctoolkit.event.zgc.*;
import com.microsoft.gctoolkit.message.ChannelName;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.message.JVMEventChannel;
import com.microsoft.gctoolkit.message.JVMEventChannelListener;
import java.io.File;
import java.util.List;
import java.util.Objects;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.util.Parsers;

public class GCCollector implements JVMEventChannel {
    private static final Logger LOG = LoggerFactory.getLogger(GCCollector.class);
    private final String path;
    private final List<DataSourceParser> parsers;

    public GCCollector(File file) {
        parsers = Parsers.findParsers(file);
        if (parsers.isEmpty()) {
            throw new UnsupportedOperationException(file.getPath());
        }
        parsers.forEach(parser -> parser.publishTo(this));
        path = file.getAbsolutePath();
        GC_COLLECT_FILES.attach(this, path).set(1);
    }

    public void receive(String message) {
        GC_LOG_LINES.attach(this, path).inc();
        for (DataSourceParser parser : parsers) {
            try {
                parser.receive(message);
            } catch (Exception ex) {
                LOG.error("{} error: {}", parser.getClass().getSimpleName(), message, ex);
            }
        }
    }

    @Override
    public void publish(ChannelName channel, JVMEvent event) {
        LOG.debug("{} publish {}", path, event);
        if (shouldIgnore(channel, event)) {
            return;
        }
        if (event instanceof GenerationalGCEvent) {
            recordGenerationalGCEvent((GenerationalGCEvent) event);
        } else if (event instanceof G1GCEvent) {
            recordG1GCEvent((G1GCEvent) event);
        } else if (event instanceof ZGCCycle) {
            recordZGCEvent((ZGCCycle) event);
        } else {
            LOG.debug("Publish unsupported event: {} ", event);
        }
    }

    @Override
    public void close() {
        detach(this);
    }

    @Override
    public void registerListener(JVMEventChannelListener listener) {
        throw new UnsupportedOperationException();
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) return true;
        if (!(other instanceof GCCollector)) return false;
        return Objects.equals(path, ((GCCollector) other).path);
    }

    @Override
    public int hashCode() {
        return Objects.hash(path);
    }

    @VisibleForTesting
    public void recordGenerationalGCEvent(GenerationalGCEvent event) {
        String category = Parsers.parseGenerationalGCEventCategory(event);
        LOG.debug("Collect GenerationalGCEvent {}", category);
        recordGCEvent(category, event.getDuration());
        if (event instanceof GenerationalGCPauseEvent) {
            recordGenerationalGCPauseEvent((GenerationalGCPauseEvent) event);
            recordGCPauseEvent(category, event.getDuration());
        }
    }

    @VisibleForTesting
    public void recordG1GCEvent(G1GCEvent event) {
        final String category = Parsers.parseG1GCEventCategory(event);
        LOG.debug("Collect G1GCEvent {} ", category);

        recordGCEvent(category, event.getDuration());
        if (event instanceof G1GCPauseEvent) {
            recordG1GCPauseEvent((G1GCPauseEvent) event);
            recordGCPauseEvent(category, event.getDuration());
        }
    }

    @VisibleForTesting
    public void recordZGCEvent(ZGCCycle event) {
        final String category = Parsers.parseZGCEventCategory(event);
        LOG.debug("Collect ZGCEvent {} ", category);

        double duration = 0;
        double pauseDuration = 0;
        double pauseMarkStartDuration = event.getPauseMarkStartDuration() / 1000;
        duration += pauseMarkStartDuration;
        pauseDuration += pauseMarkStartDuration;
        ZGC_PAUSE_MARK_START_DURATION.attach(this, path).observe(pauseMarkStartDuration);
        double concurrentMarkDuration = event.getConcurrentMarkDuration() / 1000;
        duration += concurrentMarkDuration;
        ZGC_CONCURRENT_MARK_DURATION.attach(this, path).observe(concurrentMarkDuration);
        double concurrentMarkFreeDuration = event.getConcurrentMarkFreeDuration() / 1000;
        duration += concurrentMarkFreeDuration;
        ZGC_CONCURRENT_MARK_FREE_DURATION.attach(this, path).observe(concurrentMarkFreeDuration);
        double pauseMarkEndDuration = event.getPauseMarkEndDuration() / 1000;
        duration += pauseMarkEndDuration;
        pauseDuration += pauseMarkEndDuration;
        ZGC_PAUSE_MARK_END_DURATION.attach(this, path).observe(pauseMarkEndDuration);
        double concurrentProcessNonStrongReferencesDuration =
                event.getConcurrentProcessNonStrongReferencesDuration() / 1000;
        duration += concurrentProcessNonStrongReferencesDuration;
        ZGC_PROCESS_NON_STRONG_REFERENCES_DURATION
                .attach(this, path)
                .observe(concurrentProcessNonStrongReferencesDuration);
        double concurrentResetRelocationSetDuration =
                event.getConcurrentResetRelocationSetDuration() / 1000;
        duration += concurrentResetRelocationSetDuration;
        ZGC_CONCURRENT_RESET_RELOCATIONSET_DURATION
                .attach(this, path)
                .observe(concurrentResetRelocationSetDuration);
        double concurrentSelectRelocationSetDuration =
                event.getConcurrentSelectRelocationSetDuration() / 1000;
        duration += concurrentSelectRelocationSetDuration;
        ZGC_CONCURRENT_SELECT_RELOCATIONSET_DURATION
                .attach(this, path)
                .observe(concurrentSelectRelocationSetDuration);
        double pauseRelocateStartDuration = event.getPauseRelocateStartDuration() / 1000;
        duration += pauseRelocateStartDuration;
        pauseDuration += pauseRelocateStartDuration;
        ZGC_PAUSE_RELOCATE_START_DURATION.attach(this, path).observe(pauseRelocateStartDuration);
        double concurrentRelocateDuration = event.getConcurrentRelocateDuration() / 1000;
        duration += concurrentRelocateDuration;
        ZGC_CONCURRENT_RELOCATE_DURATION.attach(this, path).observe(concurrentRelocateDuration);

        recordGCEvent(category, duration);
        recordGCPauseEvent(category, pauseDuration);

        double load1m = event.getLoadAverageAt(1);
        double load5m = event.getLoadAverageAt(5);
        double load15m = event.getLoadAverageAt(15);
        ZGC_LOAD_1m.attach(this, path).set(load1m);
        ZGC_LOAD_5m.attach(this, path).set(load5m);
        ZGC_LOAD_15m.attach(this, path).set(load15m);

        double mmu_2ms = event.getMMU(2) / 100d;
        double mmu_5ms = event.getMMU(5) / 100d;
        double mmu_10ms = event.getMMU(10) / 100d;
        double mmu_20ms = event.getMMU(20) / 100d;
        double mmu_50ms = event.getMMU(50) / 100d;
        double mmu_100ms = event.getMMU(100) / 100d;
        ZGC_MMU_2MS.attach(this, path).set(mmu_2ms);
        ZGC_MMU_5MS.attach(this, path).set(mmu_5ms);
        ZGC_MMU_10MS.attach(this, path).set(mmu_10ms);
        ZGC_MMU_20MS.attach(this, path).set(mmu_20ms);
        ZGC_MMU_50MS.attach(this, path).set(mmu_50ms);
        ZGC_MMU_100MS.attach(this, path).set(mmu_100ms);

        ZGCMemoryPoolSummary markStart = event.getMarkStart();
        if (markStart != null) {
            ZGC_MARK_START_USED.attach(this, path).set(markStart.getUsed() * 1024);
            ZGC_MARK_START_FREE.attach(this, path).set(markStart.getFree() * 1024);
            HEAP_OCCUPANCY_BEFORE_COLLECTION.attach(this, path).set(markStart.getUsed() * 1024);
            HEAP_SIZE_BEFORE_COLLECTION
                    .attach(this, path)
                    .set((markStart.getUsed() + markStart.getFree()) * 1024);
        }
        ZGCMemoryPoolSummary markEnd = event.getMarkEnd();
        if (markEnd != null) {
            ZGC_MARK_END_USED.attach(this, path).set(markEnd.getUsed() * 1024);
            ZGC_MARK_END_FREE.attach(this, path).set(markEnd.getFree() * 1024);
        }
        ZGCMemoryPoolSummary relocateStart = event.getRelocateStart();
        if (relocateStart != null) {
            ZGC_RELOCATE_START_USED.attach(this, path).set(relocateStart.getUsed() * 1024);
            ZGC_RELOCATE_START_FREE.attach(this, path).set(relocateStart.getFree() * 1024);
        }
        ZGCMemoryPoolSummary relocateEnd = event.getRelocateStart();
        if (relocateEnd != null) {
            ZGC_RELOCATE_END_USED.attach(this, path).set(relocateEnd.getUsed() * 1024);
            ZGC_RELOCATE_END_FREE.attach(this, path).set(relocateEnd.getFree() * 1024);
            HEAP_OCCUPANCY_AFTER_COLLECTION.attach(this, path).set(relocateEnd.getUsed() * 1024);
            HEAP_SIZE_AFTER_COLLECTION
                    .attach(this, path)
                    .set((relocateEnd.getUsed() + relocateEnd.getFree()) * 1024);
        }
        OccupancySummary live = event.getLive();
        if (live != null) {
            ZGC_LIVE_MARK_END.attach(this, path).set(live.getMarkEnd());
            ZGC_LIVE_RECLAIM_START.attach(this, path).set(live.getReclaimStart() * 1024);
            ZGC_LIVE_RECLAIM_END.attach(this, path).set(live.getReclaimEnd() * 1024);
        }
        OccupancySummary allocated = event.getAllocated();
        if (allocated != null) {
            ZGC_ALLOCATED_MARK_END.attach(this, path).set(allocated.getMarkEnd() * 1024);
            ZGC_ALLOCATED_RECLAIM_START.attach(this, path).set(allocated.getReclaimStart() * 1024);
            ZGC_ALLOCATED_RECLAIM_END.attach(this, path).set(allocated.getReclaimEnd() * 1024);
        }
        OccupancySummary garbage = event.getGarbage();
        if (garbage != null) {
            ZGC_GARBAGE_MARK_END.attach(this, path).set(garbage.getMarkEnd() * 1024);
            ZGC_GARBAGE_RECLAIM_START.attach(this, path).set(garbage.getReclaimStart() * 1024);
            ZGC_GARBAGE_RECLAIM_END.attach(this, path).set(garbage.getReclaimEnd() * 1024);
        }
        ReclaimSummary reclaimed = event.getReclaimed();
        if (reclaimed != null) {
            ZGC_RECLAIMED_RECLAIM_START.attach(this, path).set(reclaimed.getReclaimStart() * 1024);
            ZGC_RECLAIMED_RECLAIM_END.attach(this, path).set(reclaimed.getReclaimEnd() * 1024);
        }
        ReclaimSummary memorySummary = event.getMemorySummary();
        if (memorySummary != null) {
            ZGC_MEMORY_RECLAIM_START.attach(this, path).set(memorySummary.getReclaimStart() * 1024);
            ZGC_MEMORY_RECLAIM_END.attach(this, path).set(memorySummary.getReclaimEnd() * 1024);
        }
        ZGCMetaspaceSummary metaspace = event.getMetaspace();
        if (metaspace != null) {
            ZGC_METASPACE_USED.attach(this, path).set(metaspace.getUsed() * 1024);
            ZGC_METASPACE_COMMITTED.attach(this, path).set(metaspace.getCommitted() * 1024);
            ZGC_METASPACE_RESERVED.attach(this, path).set(metaspace.getReserved() * 1024);
        }
    }

    private boolean shouldIgnore(ChannelName channel, JVMEvent event) {
        if (channel == ChannelName.CMS_TENURED_POOL_PARSER_OUTBOX) {
            if (event instanceof InitialMark || event instanceof CMSRemark) {
                LOG.debug("Ignore CMS InitialMark or Remark due to GcToolkit bug");
                return true;
            }
        }
        return false;
    }

    private void recordGCEvent(String category, double duration) {
        GC_EVENT_DURATION.attach(this, path, category).observe(duration);
        GC_EVENT_LAST_MINUTE_DURATION.attach(this, path).observe(duration);
    }

    private void recordGCPauseEvent(String category, double duration) {
        GC_EVENT_PAUSE_DURATION.attach(this, path, category).observe(duration);
    }

    private void recordGenerationalGCPauseEvent(GenerationalGCPauseEvent event) {

        MemoryPoolSummary heapSummary = event.getHeap();
        if (heapSummary != null) {
            HEAP_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path)
                    .set(heapSummary.getOccupancyAfterCollection() * 1024);
            HEAP_SIZE_AFTER_COLLECTION
                    .attach(this, path)
                    .set(heapSummary.getSizeAfterCollection() * 1024);
            HEAP_SIZE_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(heapSummary.getSizeBeforeCollection() * 1024);
            HEAP_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(heapSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary tenuredSummary = event.getTenured();
        if (tenuredSummary != null) {
            OLD_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path)
                    .set(tenuredSummary.getOccupancyAfterCollection() * 1024);
            OLD_SIZE_AFTER_COLLECTION
                    .attach(this, path)
                    .set(tenuredSummary.getSizeAfterCollection() * 1024);
            OLD_SIZE_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(tenuredSummary.getSizeBeforeCollection() * 1024);
            OLD_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(tenuredSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary youngSummary = event.getYoung();
        if (youngSummary != null) {
            YOUNG_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path)
                    .set(youngSummary.getOccupancyAfterCollection() * 1024);
            YOUNG_SIZE_AFTER_COLLECTION
                    .attach(this, path)
                    .set(youngSummary.getSizeAfterCollection() * 1024);
            YOUNG_SIZE_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(youngSummary.getSizeBeforeCollection() * 1024);
            YOUNG_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(youngSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary permOrMetaspace = event.getPermOrMetaspace();
        if (permOrMetaspace != null) {
            METASPACE_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path)
                    .set(permOrMetaspace.getOccupancyAfterCollection() * 1024);
            METASPACE_SIZE_AFTER_COLLECTION
                    .attach(this, path)
                    .set(permOrMetaspace.getSizeAfterCollection() * 1024);
            METASPACE_SIZE_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(permOrMetaspace.getSizeBeforeCollection() * 1024);
            METASPACE_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(permOrMetaspace.getOccupancyBeforeCollection() * 1024);
        }

        if (event instanceof CMSPhase) {
            double classUnloadingProcessingTime = event.getClassUnloadingProcessingTime();
            double symbolTableProcessingTime = event.getSymbolTableProcessingTime();
            double stringTableProcessingTime = event.getStringTableProcessingTime();
            double symbolAndStringTableProcessingTime =
                    event.getSymbolAndStringTableProcessingTime();
            if (classUnloadingProcessingTime > 0.0) {
                CMS_CLASS_UNLOADING_PROCESS_TIME
                        .attach(this, path)
                        .observe(classUnloadingProcessingTime);
            }
            if (symbolTableProcessingTime > 0.0) {
                CMS_SYMBOL_TABLE_PROCESS_TIME.attach(this, path).observe(symbolTableProcessingTime);
            }
            if (stringTableProcessingTime > 0.0) {
                CMS_STRING_TABLE_PROCESS_TIME.attach(this, path).observe(stringTableProcessingTime);
            }
            if (symbolAndStringTableProcessingTime > 0.0) {
                CMS_SYMBOL_AND_STRING_TABLE_PROCESS_TIME
                        .attach(this, path)
                        .observe(symbolAndStringTableProcessingTime);
            }
        }

        ReferenceGCSummary referenceGCSummary = event.getReferenceGCSummary();
        recordReferenceSummary(referenceGCSummary);
    }

    private void recordReferenceSummary(ReferenceGCSummary referenceGCSummary) {
        if (referenceGCSummary != null) {
            int softReferenceCount = referenceGCSummary.getSoftReferenceCount();
            SOFT_REFERENCE_COUNT.attach(this, path).set(softReferenceCount);
            double softReferencePauseTime = referenceGCSummary.getSoftReferencePauseTime();
            SOFT_REFERENCE_PAUSE_TIME.attach(this, path).observe(softReferencePauseTime);

            int weakReferenceCount = referenceGCSummary.getWeakReferenceCount();
            WEAK_REFERENCE_COUNT.attach(this, path).set(weakReferenceCount);
            double weakReferencePauseTime = referenceGCSummary.getWeakReferencePauseTime();
            WEAK_REFERENCE_PAUSE_TIME.attach(this, path).observe(weakReferencePauseTime);

            int finalReferenceCount = referenceGCSummary.getFinalReferenceCount();
            FINAL_REFERENCE_COUNT.attach(this, path).set(finalReferenceCount);
            double finalReferencePauseTime = referenceGCSummary.getFinalReferencePauseTime();
            FINAL_REFERENCE_PAUSE_TIME.attach(this, path).observe(finalReferencePauseTime);

            int phantomReferenceCount = referenceGCSummary.getPhantomReferenceCount();
            PHANTOM_REFERENCE_COUNT.attach(this, path).set(phantomReferenceCount);
            int phantomReferenceFreedCount = referenceGCSummary.getPhantomReferenceFreedCount();
            PHANTOM_REFERENCE_FREE_COUNT.attach(this, path).set(phantomReferenceFreedCount);
            double phantomReferencePauseTime = referenceGCSummary.getPhantomReferencePauseTime();
            PHANTOM_REFERENCE_PAUSE_TIME.attach(this, path).observe(phantomReferencePauseTime);

            int jniWeakReferenceCount = referenceGCSummary.getJniWeakReferenceCount();
            JNI_WEAK_REFERENCE_COUNT.attach(this, path).set(jniWeakReferenceCount);
            double jniWeakReferencePauseTime = referenceGCSummary.getJniWeakReferencePauseTime();
            JNI_WEAK_REFERENCE_PAUSE_TIME.attach(this, path).observe(jniWeakReferencePauseTime);
        }
    }

    private void recordG1GCPauseEvent(G1GCPauseEvent event) {

        MemoryPoolSummary heapSummary = event.getHeap();
        if (heapSummary != null) {
            HEAP_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path)
                    .set(heapSummary.getOccupancyAfterCollection() * 1024);
            HEAP_SIZE_AFTER_COLLECTION
                    .attach(this, path)
                    .set(heapSummary.getSizeAfterCollection() * 1024);
            HEAP_SIZE_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(heapSummary.getSizeBeforeCollection() * 1024);
            HEAP_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path)
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
                        .attach(this, path)
                        .set(eden.getOccupancyAfterCollection() * 1024);
                youngOccupancyAfterCollection += eden.getOccupancyAfterCollection() * 1024;
            }
            if (eden.getOccupancyBeforeCollection() >= 0) {
                G1_EDEN_OCCUPANCY_BEFORE_COLLECTION
                        .attach(this, path)
                        .set(eden.getOccupancyBeforeCollection() * 1024);
                youngOccupancyBeforeCollection += eden.getOccupancyBeforeCollection() * 1024;
            }
            if (eden.getSizeBeforeCollection() > 0) {
                G1_EDEN_SIZE_BEFORE_COLLECTION
                        .attach(this, path)
                        .set(eden.getSizeBeforeCollection() * 1024);
                youngSizeBeforeCollection += eden.getSizeBeforeCollection() * 1024;
            }
            if (eden.getSizeAfterCollection() > 0) {
                G1_EDEN_SIZE_AFTER_COLLECTION
                        .attach(this, path)
                        .set(eden.getSizeAfterCollection() * 1024);
                youngSizeAfterCollection += eden.getSizeAfterCollection() * 1024;
            }
        }

        SurvivorMemoryPoolSummary survivor = event.getSurvivor();
        if (survivor != null) {
            if (survivor.getOccupancyAfterCollection() >= 0) {
                G1_SURVIVOR_HEAP_OCCUPANCY_AFTER_COLLECTION
                        .attach(this, path)
                        .set(survivor.getOccupancyAfterCollection() * 1024);
                youngOccupancyAfterCollection += survivor.getOccupancyAfterCollection() * 1024;
            }
            if (survivor.getOccupancyBeforeCollection() >= 0) {
                G1_SURVIVOR_HEAP_OCCUPANCY_BEFORE_COLLECTION
                        .attach(this, path)
                        .set(survivor.getOccupancyBeforeCollection() * 1024);

                youngOccupancyBeforeCollection += survivor.getOccupancyBeforeCollection() * 1024;
            }
            if (survivor.getSize() > 0) {
                G1_SURVIVOR_SIZE.attach(this, path).set(survivor.getSize() * 1024);
                youngSizeBeforeCollection += survivor.getSize() * 1024;
                youngSizeAfterCollection += survivor.getSize() * 1024;
            }
        }

        // young generation
        if (youngOccupancyBeforeCollection >= 0) {
            YOUNG_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(youngOccupancyBeforeCollection);
        }
        if (youngOccupancyAfterCollection >= 0) {
            YOUNG_OCCUPANCY_AFTER_COLLECTION.attach(this, path).set(youngOccupancyAfterCollection);
        }
        if (youngSizeBeforeCollection >= 0) {
            YOUNG_SIZE_BEFORE_COLLECTION.attach(this, path).set(youngSizeBeforeCollection);
        }
        if (youngSizeAfterCollection >= 0) {
            YOUNG_SIZE_AFTER_COLLECTION.attach(this, path).set(youngSizeAfterCollection);
        }

        MemoryPoolSummary metaspace = event.getPermOrMetaspace();
        if (metaspace != null) {
            METASPACE_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path)
                    .set(metaspace.getOccupancyAfterCollection() * 1024);
            METASPACE_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(metaspace.getOccupancyBeforeCollection() * 1024);
            METASPACE_SIZE_BEFORE_COLLECTION
                    .attach(this, path)
                    .set(metaspace.getSizeBeforeCollection() * 1024);
            METASPACE_SIZE_AFTER_COLLECTION
                    .attach(this, path)
                    .set(metaspace.getSizeAfterCollection() * 1024);
        }

        ReferenceGCSummary referenceGCSummary = event.getReferenceGCSummary();
        recordReferenceSummary(referenceGCSummary);

        RegionSummary edenRegion = event.getEdenRegionSummary();
        if (edenRegion != null) {
            if (edenRegion.getBefore() >= 0) {
                G1_EDEN_REGION_BEFORE.attach(this, path).set(edenRegion.getBefore());
            }
            if (edenRegion.getAfter() >= 0) {
                G1_EDEN_REGION_AFTER.attach(this, path).set(edenRegion.getAfter());
            }
            if (edenRegion.getAssigned() >= 0) {
                G1_EDEN_REGION_ASSIGN.attach(this, path).set(edenRegion.getAssigned());
            }
        }
        RegionSummary survivorRegion = event.getSurvivorRegionSummary();
        if (survivorRegion != null) {
            if (survivorRegion.getBefore() >= 0) {
                G1_SURVIVOR_REGION_BEFORE.attach(this, path).set(survivorRegion.getBefore());
            }
            if (survivorRegion.getAfter() >= 0) {
                G1_SURVIVOR_REGION_AFTER.attach(this, path).set(survivorRegion.getAfter());
            }
            if (survivorRegion.getAssigned() >= 0) {
                G1_SURVIVOR_REGION_ASSIGN.attach(this, path).set(survivorRegion.getAssigned());
            }
        }
        RegionSummary oldRegion = event.getOldRegionSummary();
        if (oldRegion != null) {
            if (oldRegion.getBefore() >= 0) {
                G1_OLD_REGION_BEFORE.attach(this, path).set(oldRegion.getBefore());
            }
            if (oldRegion.getAfter() >= 0) {
                G1_OLD_REGION_AFTER.attach(this, path).set(oldRegion.getAfter());
            }
            if (oldRegion.getAssigned() >= 0) {
                G1_OLD_REGION_ASSIGN.attach(this, path).set(oldRegion.getAssigned());
            }
        }
        RegionSummary humongousRegion = event.getHumongousRegionSummary();
        if (humongousRegion != null) {
            if (humongousRegion.getBefore() >= 0) {
                G1_HUMONGOUS_REGION_BEFORE.attach(this, path).set(humongousRegion.getBefore());
            }
            if (humongousRegion.getAfter() >= 0) {
                G1_HUMONGOUS_REGION_AFTER.attach(this, path).set(humongousRegion.getAfter());
            }
            if (humongousRegion.getAssigned() >= 0) {
                G1_HUMONGOUS_REGION_ASSIGN.attach(this, path).set(humongousRegion.getAssigned());
            }
        }
        RegionSummary archiveRegion = event.getArchiveRegionSummary();
        if (archiveRegion != null) {
            if (archiveRegion.getBefore() >= 0) {
                G1_ARCHIVE_REGION_BEFORE.attach(this, path).set(archiveRegion.getBefore());
            }
            if (archiveRegion.getAfter() >= 0) {
                G1_ARCHIVE_REGION_AFTER.attach(this, path).set(archiveRegion.getAfter());
            }
            if (archiveRegion.getAssigned() >= 0) {
                G1_ARCHIVE_REGION_ASSIGN.attach(this, path).set(archiveRegion.getAssigned());
            }
        }
    }
}
