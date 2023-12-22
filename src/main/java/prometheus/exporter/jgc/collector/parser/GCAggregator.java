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
package prometheus.exporter.jgc.collector.parser;

import static java.lang.Class.forName;
import static prometheus.exporter.jgc.collector.parser.Metrics.*;

import com.google.common.annotations.VisibleForTesting;
import com.microsoft.gctoolkit.event.*;
import com.microsoft.gctoolkit.event.g1gc.*;
import com.microsoft.gctoolkit.event.generational.*;
import com.microsoft.gctoolkit.event.jvm.JVMEvent;
import com.microsoft.gctoolkit.event.jvm.Safepoint;
import com.microsoft.gctoolkit.event.jvm.SurvivorRecord;
import com.microsoft.gctoolkit.event.zgc.*;
import com.microsoft.gctoolkit.io.SingleGCLogFile;
import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.message.ChannelName;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.message.JVMEventChannel;
import com.microsoft.gctoolkit.message.JVMEventChannelListener;
import io.prometheus.client.*;
import java.io.File;
import java.lang.reflect.InvocationTargetException;
import java.util.*;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.collector.CleanableCollectorRegistry;

public class GCAggregator implements JVMEventChannel {
    private static final Logger LOG = LoggerFactory.getLogger(GCAggregator.class);

    private static final String[] DEFAULT_PARSERS = {
        "com.microsoft.gctoolkit.parser.CMSTenuredPoolParser",
        "com.microsoft.gctoolkit.parser.GenerationalHeapParser",
        "com.microsoft.gctoolkit.parser.JVMEventParser",
        "com.microsoft.gctoolkit.parser.PreUnifiedG1GCParser",
        "com.microsoft.gctoolkit.parser.ShenandoahParser",
        "com.microsoft.gctoolkit.parser.SurvivorMemoryPoolParser",
        "com.microsoft.gctoolkit.parser.UnifiedG1GCParser",
        "com.microsoft.gctoolkit.parser.UnifiedGenerationalParser",
        "com.microsoft.gctoolkit.parser.UnifiedJVMEventParser",
        "com.microsoft.gctoolkit.parser.UnifiedSurvivorMemoryPoolParser",
        "com.microsoft.gctoolkit.parser.ZGCParser"
    };

    private final String path;
    private final List<DataSourceParser> parsers;

    public GCAggregator(File file) {
        this.path = file.getPath();
        this.parsers = loadParsers(file);
        this.parsers.forEach(parser -> parser.publishTo(this));
    }

    public void receive(String message) {
        GC_LOG_LINES.labels(path).inc();
        for (DataSourceParser parser : parsers) {
            try {
                parser.receive(message);
            } catch (Exception ex) {
                LOG.error("parse error: {}", message, ex);
            }
        }
    }

    @Override
    public void publish(ChannelName channel, JVMEvent event) {
        LOG.debug("{} occurs {}", path, event.getClass().getSimpleName());
        if (event instanceof GenerationalGCEvent) {
            recordGenerationalGCEvent((GenerationalGCEvent) event);
        } else if (event instanceof G1GCEvent) {
            recordG1GCEvent((G1GCEvent) event);
        } else if (event instanceof ZGCCycle) {
            recordZGCEvent((ZGCCycle) event);
        } else if (event instanceof Safepoint) {
            recordSafePoint((Safepoint) event);
        } else if (event instanceof SurvivorRecord) {
            recordSurvivorRecord((SurvivorRecord) event);
        } else {
            LOG.warn("unsupported JvmEvent: {} ", event);
        }
    }

    @Override
    public void close() {
        CleanableCollectorRegistry.DEFAULT.clean(
                sample -> {
                    int index = sample.labelNames.indexOf("path");
                    if (index == -1) {
                        return false;
                    }
                    return path.equals(sample.labelValues.get(index));
                });
        LOG.info("clean {} metrics", path);
    }

    @Override
    public void registerListener(JVMEventChannelListener listener) {}

    @VisibleForTesting
    public void recordGenerationalGCEvent(GenerationalGCEvent event) {
        String category = parseGCEventCategory(event);
        LOG.debug("{} Collect GenerationalGCEvent {}", event.getClass(), category);
        recordGCEvent(category, event.getDuration());
        if (event instanceof GenerationalGCPauseEvent) {
            recordGenerationalGCPauseEvent((GenerationalGCPauseEvent) event);
            recordGCPauseEvent(category, event.getDuration());
        } else if (event instanceof CMSConcurrentEvent) {
            recordCMSConcurrentEvent((CMSConcurrentEvent) event);
        }
    }

    @VisibleForTesting
    public void recordG1GCEvent(G1GCEvent event) {
        final String category = parseGCEventCategory(event);
        LOG.debug("{} Collect G1GCEvent {} ", event.getClass(), category);
        recordGCEvent(category, event.getDuration());
        if (event instanceof G1GCPauseEvent) {
            recordG1GCPauseEvent((G1GCPauseEvent) event);
            recordGCPauseEvent(category, event.getDuration());
        }
    }

    @VisibleForTesting
    public void recordZGCEvent(ZGCCycle event) {
        LOG.debug("Collect ZGCEvent");
        final String category = parseGCEventCategory(event);

        double duration = 0;
        double pauseDuration = 0;
        double pauseMarkStartDuration = event.getPauseMarkStartDuration() / 1000;
        duration += pauseMarkStartDuration;
        pauseDuration += pauseMarkStartDuration;
        ZGC_PAUSE_MARK_START_DURATION.labels(path).observe(pauseMarkStartDuration);
        double concurrentMarkDuration = event.getConcurrentMarkDuration() / 1000;
        duration += concurrentMarkDuration;
        ZGC_CONCURRENT_MARK_DURATION.labels(path).observe(concurrentMarkDuration);
        double concurrentMarkFreeDuration = event.getConcurrentMarkFreeDuration() / 1000;
        duration += concurrentMarkFreeDuration;
        ZGC_CONCURRENT_MARK_FREE_DURATION.labels(path).observe(concurrentMarkFreeDuration);
        double pauseMarkEndDuration = event.getPauseMarkEndDuration() / 1000;
        duration += pauseMarkEndDuration;
        pauseDuration += pauseMarkEndDuration;
        ZGC_PAUSE_MARK_END_DURATION.labels(path).observe(pauseMarkEndDuration);
        double concurrentProcessNonStrongReferencesDuration =
                event.getConcurrentProcessNonStrongReferencesDuration() / 1000;
        duration += concurrentProcessNonStrongReferencesDuration;
        ZGC_PROCESS_NON_STRONG_REFERENCES_DURATION
                .labels(path)
                .observe(concurrentProcessNonStrongReferencesDuration);
        double concurrentResetRelocationSetDuration =
                event.getConcurrentResetRelocationSetDuration() / 1000;
        duration += concurrentResetRelocationSetDuration;
        ZGC_CONCURRENT_RESET_RELOCATIONSET_DURATION
                .labels(path)
                .observe(concurrentResetRelocationSetDuration);
        double concurrentSelectRelocationSetDuration =
                event.getConcurrentSelectRelocationSetDuration() / 1000;
        duration += concurrentSelectRelocationSetDuration;
        ZGC_CONCURRENT_SELECT_RELOCATIONSET_DURATION
                .labels(path)
                .observe(concurrentSelectRelocationSetDuration);
        double pauseRelocateStartDuration = event.getPauseRelocateStartDuration() / 1000;
        duration += pauseRelocateStartDuration;
        pauseDuration += pauseRelocateStartDuration;
        ZGC_PAUSE_RELOCATE_START_DURATION.labels(path).observe(pauseRelocateStartDuration);
        double concurrentRelocateDuration = event.getConcurrentRelocateDuration() / 1000;
        duration += concurrentRelocateDuration;
        ZGC_CONCURRENT_RELOCATE_DURATION.labels(path).observe(concurrentRelocateDuration);

        recordGCEvent(category, duration);
        recordGCPauseEvent(category, pauseDuration);

        double load1m = event.getLoadAverageAt(1);
        double load5m = event.getLoadAverageAt(5);
        double load15m = event.getLoadAverageAt(15);
        ZGC_LOAD_1m.labels(path).set(load1m);
        ZGC_LOAD_5m.labels(path).set(load5m);
        ZGC_LOAD_15m.labels(path).set(load15m);

        double mmu_2ms = event.getMMU(2) / 100d;
        double mmu_5ms = event.getMMU(5) / 100d;
        double mmu_10ms = event.getMMU(10) / 100d;
        double mmu_20ms = event.getMMU(20) / 100d;
        double mmu_50ms = event.getMMU(50) / 100d;
        double mmu_100ms = event.getMMU(100) / 100d;
        ZGC_MMU_2MS.labels(path).set(mmu_2ms);
        ZGC_MMU_5MS.labels(path).set(mmu_5ms);
        ZGC_MMU_10MS.labels(path).set(mmu_10ms);
        ZGC_MMU_20MS.labels(path).set(mmu_20ms);
        ZGC_MMU_50MS.labels(path).set(mmu_50ms);
        ZGC_MMU_100MS.labels(path).set(mmu_100ms);

        ZGCMemoryPoolSummary markStart = event.getMarkStart();
        if (markStart != null) {
            ZGC_MARK_START_USED.labels(path).set(markStart.getUsed() * 1024);
            ZGC_MARK_START_FREE.labels(path).set(markStart.getFree() * 1024);
        }
        ZGCMemoryPoolSummary markEnd = event.getMarkEnd();
        if (markEnd != null) {
            ZGC_MARK_END_USED.labels(path).set(markEnd.getUsed() * 1024);
            ZGC_MARK_END_FREE.labels(path).set(markEnd.getFree() * 1024);
        }
        ZGCMemoryPoolSummary relocateStart = event.getRelocateStart();
        if (relocateStart != null) {
            ZGC_RELOCATE_START_USED.labels(path).set(relocateStart.getUsed() * 1024);
            ZGC_RELOCATE_START_FREE.labels(path).set(relocateStart.getFree() * 1024);
        }
        ZGCMemoryPoolSummary relocateEnd = event.getRelocateStart();
        if (relocateEnd != null) {
            ZGC_RELOCATE_END_USED.labels(path).set(relocateEnd.getUsed() * 1024);
            ZGC_RELOCATE_END_FREE.labels(path).set(relocateEnd.getFree() * 1024);
        }
        OccupancySummary live = event.getLive();
        if (live != null) {
            ZGC_LIVE_MARK_END.labels(path).set(live.getMarkEnd());
            ZGC_LIVE_RECLAIM_START.labels(path).set(live.getReclaimStart() * 1024);
            ZGC_LIVE_RECLAIM_END.labels(path).set(live.getReclaimEnd() * 1024);
        }
        OccupancySummary allocated = event.getAllocated();
        if (allocated != null) {
            ZGC_ALLOCATED_MARK_END.labels(path).set(allocated.getMarkEnd() * 1024);
            ZGC_ALLOCATED_RECLAIM_START.labels(path).set(allocated.getReclaimStart() * 1024);
            ZGC_ALLOCATED_RECLAIM_END.labels(path).set(allocated.getReclaimEnd() * 1024);
        }
        OccupancySummary garbage = event.getGarbage();
        if (garbage != null) {
            ZGC_GARBAGE_MARK_END.labels(path).set(garbage.getMarkEnd() * 1024);
            ZGC_GARBAGE_RECLAIM_START.labels(path).set(garbage.getReclaimStart() * 1024);
            ZGC_GARBAGE_RECLAIM_END.labels(path).set(garbage.getReclaimEnd() * 1024);
        }
        ReclaimSummary reclaimed = event.getReclaimed();
        if (reclaimed != null) {
            ZGC_RECLAIMED_RECLAIM_START.labels(path).set(reclaimed.getReclaimStart() * 1024);
            ZGC_RECLAIMED_RECLAIM_END.labels(path).set(reclaimed.getReclaimEnd() * 1024);
        }
        ReclaimSummary memorySummary = event.getMemorySummary();
        if (memorySummary != null) {
            ZGC_MEMORY_RECLAIM_START.labels(path).set(memorySummary.getReclaimStart() * 1024);
            ZGC_MEMORY_RECLAIM_END.labels(path).set(memorySummary.getReclaimEnd() * 1024);
        }
        ZGCMetaspaceSummary metaspace = event.getMetaspace();
        if (metaspace != null) {
            ZGC_METASPACE_USED.labels(path).set(metaspace.getUsed() * 1024);
            ZGC_METASPACE_COMMITTED.labels(path).set(metaspace.getCommitted() * 1024);
            ZGC_METASPACE_RESERVED.labels(path).set(metaspace.getReserved() * 1024);
        }
    }

    private void recordSafePoint(Safepoint safepoint) {
        LOG.debug("{} Collect Safepoint", path);
        recordGCEvent("safepoint", 0d);
        int totalNumberOfApplicationThreads = safepoint.getTotalNumberOfApplicationThreads();
        SAFEPOINT_TOTAL_NUMBER_OF_APPLICATION_THREADS
                .labels(path)
                .set(totalNumberOfApplicationThreads);
        int initiallyRunning = safepoint.getInitiallyRunning();
        SAFEPOINT_INITIALLY_RUNNING.labels(path).set(initiallyRunning);
        int waitingToBlock = safepoint.getWaitingToBlock();
        SAFEPOINT_WAITING_TO_BLOCK.labels(path).set(waitingToBlock);
        int spinDuration = safepoint.getSpinDuration();
        SAFEPOINT_SPIN_DURATION.labels(path).observe(spinDuration);
        int blockDuration = safepoint.getBlockDuration();
        SAFEPOINT_BLOCK_DURATION.labels(path).observe(blockDuration);
        int syncDuration = safepoint.getSyncDuration();
        SAFEPOINT_SYNC_DURATION.labels(path).observe(syncDuration);
        int cleanupDuration = safepoint.getCleanupDuration();
        SAFEPOINT_CLEANUP_DURATION.labels(path).observe(cleanupDuration);
        int vmopDuration = safepoint.getVmopDuration();
        SAFEPOINT_VMOP_DURATION.labels(path).observe(vmopDuration);
    }

    private void recordSurvivorRecord(SurvivorRecord record) {
        LOG.debug("{} Collect SurvivorRecord", path);
        recordGCEvent("survivor", 0d);
        long desiredOccupancyAfterCollection = record.getDesiredOccupancyAfterCollection();
        SURVIVOR_DESIRED_OCCUPANCY_AFTER_COLLECTION
                .labels(path)
                .set(desiredOccupancyAfterCollection);
        int calculatedTenuringThreshold = record.getCalculatedTenuringThreshold();
        SURVIVOR_CALCULATED_TENURING_THRESHOLD.labels(path).set(calculatedTenuringThreshold);
        int maxTenuringThreshold = record.getMaxTenuringThreshold();
        SURVIVOR_MAX_TENURING_THRESHOLD.labels(path).set(maxTenuringThreshold);
    }

    private void recordGCEvent(String category, double duration) {
        GC_EVENT_DURATION.labels(path, category).observe(duration);
    }

    private void recordGCPauseEvent(String category, double duration) {
        GC_EVENT_PAUSE_DURATION.labels(path, category).observe(duration);
    }

    private void recordCMSConcurrentEvent(CMSConcurrentEvent event) {
        USER_CPU_TIME.labels(path).observe(event.getCpuTime());
        REAL_CPU_TIME.labels(path).observe(event.getWallClockTime());
    }

    private void recordGenerationalGCPauseEvent(GenerationalGCPauseEvent event) {

        CPUSummary cpuSummary = event.getCpuSummary();
        if (cpuSummary != null) {
            SYS_CPU_TIME.labels(path).observe(cpuSummary.getKernel());
            USER_CPU_TIME.labels(path).observe(cpuSummary.getUser());
            REAL_CPU_TIME.labels(path).observe(cpuSummary.getWallClock());
        }

        MemoryPoolSummary heapSummary = event.getHeap();
        if (heapSummary != null) {
            GENERATIONAL_HEAP_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(heapSummary.getOccupancyAfterCollection() * 1024);
            GENERATIONAL_HEAP_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(heapSummary.getSizeAfterCollection() * 1024);
            GENERATIONAL_HEAP_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(heapSummary.getSizeBeforeCollection() * 1024);
            GENERATIONAL_HEAP_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(heapSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary tenuredSummary = event.getTenured();
        if (tenuredSummary != null) {
            GENERATIONAL_TENURED_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(tenuredSummary.getOccupancyAfterCollection() * 1024);
            GENERATIONAL_TENURED_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(tenuredSummary.getSizeAfterCollection() * 1024);
            GENERATIONAL_TENURED_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(tenuredSummary.getSizeBeforeCollection() * 1024);
            GENERATIONAL_TENURED_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(tenuredSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary youngSummary = event.getYoung();
        if (youngSummary != null) {
            GENERATIONAL_YOUNG_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(youngSummary.getOccupancyAfterCollection() * 1024);
            GENERATIONAL_YOUNG_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(youngSummary.getSizeAfterCollection() * 1024);
            GENERATIONAL_YOUNG_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(youngSummary.getSizeBeforeCollection() * 1024);
            GENERATIONAL_YOUNG_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(youngSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary classSummary = event.getClassspace();
        if (classSummary != null) {
            GENERATIONAL_CLASSSPACE_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(classSummary.getOccupancyAfterCollection() * 1024);
            GENERATIONAL_CLASSSPACE_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(classSummary.getSizeAfterCollection() * 1024);
            GENERATIONAL_CLASSSPACE_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(classSummary.getSizeBeforeCollection() * 1024);
            GENERATIONAL_CLASSSPACE_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(classSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary nonClassSummary = event.getNonClassspace();
        if (nonClassSummary != null) {
            GENERATIONAL_NONCLASSSPACE_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(nonClassSummary.getOccupancyAfterCollection() * 1024);
            GENERATIONAL_NONCLASSSPACE_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(nonClassSummary.getSizeAfterCollection() * 1024);
            GENERATIONAL_NONCLASSSPACE_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(nonClassSummary.getSizeBeforeCollection() * 1024);
            GENERATIONAL_NONCLASSSPACE_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(nonClassSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary permOrMetaspace = event.getPermOrMetaspace();
        if (permOrMetaspace != null) {
            GENERATIONAL_METASPACE_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(permOrMetaspace.getOccupancyAfterCollection() * 1024);
            GENERATIONAL_METASPACE_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(permOrMetaspace.getSizeAfterCollection() * 1024);
            GENERATIONAL_METASPACE_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(permOrMetaspace.getSizeBeforeCollection() * 1024);
            GENERATIONAL_METASPACE_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(permOrMetaspace.getOccupancyBeforeCollection() * 1024);
        }

        double classUnloadingProcessingTime = event.getClassUnloadingProcessingTime();
        double symbolTableProcessingTime = event.getSymbolTableProcessingTime();
        double stringTableProcessingTime = event.getStringTableProcessingTime();
        double symbolAndStringTableProcessingTime = event.getSymbolAndStringTableProcessingTime();
        if (classUnloadingProcessingTime > 0.0) {
            GENERATIONAL_CLASS_UNLOADING_PROCESS_TIME
                    .labels(path)
                    .observe(classUnloadingProcessingTime);
        }
        if (symbolTableProcessingTime > 0.0) {
            GENERATIONAL_SYMBOL_TABLE_PROCESS_TIME.labels(path).observe(symbolTableProcessingTime);
        }
        if (stringTableProcessingTime > 0.0) {
            GENERATIONAL_STRING_TABLE_PROCESS_TIME.labels(path).observe(stringTableProcessingTime);
        }
        if (symbolAndStringTableProcessingTime > 0.0) {
            GENERATIONAL_SYMBOL_AND_STRING_TABLE_PROCESS_TIME
                    .labels(path)
                    .observe(symbolAndStringTableProcessingTime);
        }

        ReferenceGCSummary referenceGCSummary = event.getReferenceGCSummary();
        if (referenceGCSummary != null) {
            int softReferenceCount = referenceGCSummary.getSoftReferenceCount();
            GENERATIONAL_SOFT_REFERENCE_COUNT.labels(path).set(softReferenceCount);
            double softReferencePauseTime = referenceGCSummary.getSoftReferencePauseTime();
            GENERATIONAL_SOFT_REFERENCE_PAUSE_TIME.labels(path).observe(softReferencePauseTime);

            int weakReferenceCount = referenceGCSummary.getWeakReferenceCount();
            GENERATIONAL_WEAK_REFERENCE_COUNT.labels(path).set(weakReferenceCount);
            double weakReferencePauseTime = referenceGCSummary.getWeakReferencePauseTime();
            GENERATIONAL_WEAK_REFERENCE_PAUSE_TIME.labels(path).observe(weakReferencePauseTime);

            int finalReferenceCount = referenceGCSummary.getFinalReferenceCount();
            GENERATIONAL_FINAL_REFERENCE_COUNT.labels(path).set(finalReferenceCount);
            double finalReferencePauseTime = referenceGCSummary.getFinalReferencePauseTime();
            GENERATIONAL_FINAL_REFERENCE_PAUSE_TIME.labels(path).observe(finalReferencePauseTime);

            int phantomReferenceCount = referenceGCSummary.getPhantomReferenceCount();
            GENERATIONAL_PHANTOM_REFERENCE_COUNT.labels(path).set(phantomReferenceCount);
            int phantomReferenceFreedCount = referenceGCSummary.getPhantomReferenceFreedCount();
            GENERATIONAL_PHANTOM_REFERENCE_FREE_COUNT.labels(path).set(phantomReferenceFreedCount);
            double phantomReferencePauseTime = referenceGCSummary.getPhantomReferencePauseTime();
            GENERATIONAL_PHANTOM_REFERENCE_PAUSE_TIME
                    .labels(path)
                    .observe(phantomReferencePauseTime);

            int jniWeakReferenceCount = referenceGCSummary.getJniWeakReferenceCount();
            GENERATIONAL_JNI_WEAK_REFERENCE_COUNT.labels(path).set(jniWeakReferenceCount);
            double jniWeakReferencePauseTime = referenceGCSummary.getJniWeakReferencePauseTime();
            GENERATIONAL_JNI_WEAK_REFERENCE_PAUSE_TIME
                    .labels(path)
                    .observe(jniWeakReferencePauseTime);
        }
    }

    private void recordG1GCPauseEvent(G1GCPauseEvent event) {

        CPUSummary cpuSummary = event.getCpuSummary();
        if (cpuSummary != null) {
            SYS_CPU_TIME.labels(path).observe(cpuSummary.getKernel());
            USER_CPU_TIME.labels(path).observe(cpuSummary.getUser());
            REAL_CPU_TIME.labels(path).observe(cpuSummary.getWallClock());
        }

        MemoryPoolSummary heapSummary = event.getHeap();
        if (heapSummary != null) {
            G1_HEAP_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(heapSummary.getOccupancyAfterCollection() * 1024);
            G1_HEAP_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(heapSummary.getSizeAfterCollection() * 1024);
            G1_HEAP_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(heapSummary.getSizeBeforeCollection() * 1024);
            G1_HEAP_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(heapSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary eden = event.getEden();
        if (eden != null) {
            if (eden.getOccupancyAfterCollection() >= 0) {
                G1_EDEN_OCCUPANCY_AFTER_COLLECTION
                        .labels(path)
                        .set(eden.getOccupancyAfterCollection() * 1024);
            }
            if (eden.getOccupancyBeforeCollection() >= 0) {
                G1_EDEN_OCCUPANCY_BEFORE_COLLECTION
                        .labels(path)
                        .set(eden.getOccupancyBeforeCollection() * 1024);
            }
            if (eden.getSizeBeforeCollection() > 0) {
                G1_EDEN_SIZE_BEFORE_COLLECTION
                        .labels(path)
                        .set(eden.getSizeBeforeCollection() * 1024);
            }
            if (eden.getSizeAfterCollection() > 0) {
                G1_EDEN_SIZE_AFTER_COLLECTION
                        .labels(path)
                        .set(eden.getSizeAfterCollection() * 1024);
            }
        }

        SurvivorMemoryPoolSummary survivor = event.getSurvivor();
        if (survivor != null) {
            if (survivor.getOccupancyAfterCollection() >= 0) {
                G1_SURVIVOR_HEAP_OCCUPANCY_AFTER_COLLECTION
                        .labels(path)
                        .set(survivor.getOccupancyAfterCollection() * 1024);
            }
            if (survivor.getOccupancyBeforeCollection() >= 0) {
                G1_SURVIVOR_HEAP_OCCUPANCY_BEFORE_COLLECTION
                        .labels(path)
                        .set(survivor.getOccupancyBeforeCollection() * 1024);
            }
            if (survivor.getSize() > 0) {
                G1_SURVIVOR_SIZE.labels(path).set(survivor.getSize() * 1024);
            }
        }

        MemoryPoolSummary metaspace = event.getPermOrMetaspace();
        if (metaspace != null) {
            G1_META_SPACE_OCCUPANCY_AFTER_COLLECTION
                    .labels(path)
                    .set(metaspace.getOccupancyAfterCollection() * 1024);
            G1_META_SPACE_OCCUPANCY_BEFORE_COLLECTION
                    .labels(path)
                    .set(metaspace.getOccupancyBeforeCollection() * 1024);
            G1_META_SPACE_SIZE_BEFORE_COLLECTION
                    .labels(path)
                    .set(metaspace.getSizeBeforeCollection() * 1024);
            G1_META_SPACE_SIZE_AFTER_COLLECTION
                    .labels(path)
                    .set(metaspace.getSizeAfterCollection() * 1024);
        }

        ReferenceGCSummary referenceGCSummary = event.getReferenceGCSummary();
        if (referenceGCSummary != null) {
            int softReferenceCount = referenceGCSummary.getSoftReferenceCount();
            G1_SOFT_REFERENCE_COUNT.labels(path).set(softReferenceCount);
            double softReferencePauseTime = referenceGCSummary.getSoftReferencePauseTime();
            G1_SOFT_REFERENCE_PAUSE_TIME.labels(path).observe(softReferencePauseTime);

            int weakReferenceCount = referenceGCSummary.getWeakReferenceCount();
            G1_WEAK_REFERENCE_COUNT.labels(path).set(weakReferenceCount);
            double weakReferencePauseTime = referenceGCSummary.getWeakReferencePauseTime();
            G1_WEAK_REFERENCE_PAUSE_TIME.labels(path).observe(weakReferencePauseTime);

            int finalReferenceCount = referenceGCSummary.getFinalReferenceCount();
            G1_FINAL_REFERENCE_COUNT.labels(path).set(finalReferenceCount);
            double finalReferencePauseTime = referenceGCSummary.getFinalReferencePauseTime();
            G1_FINAL_REFERENCE_PAUSE_TIME.labels(path).observe(finalReferencePauseTime);

            int phantomReferenceCount = referenceGCSummary.getPhantomReferenceCount();
            G1_PHANTOM_REFERENCE_COUNT.labels(path).set(phantomReferenceCount);
            int phantomReferenceFreedCount = referenceGCSummary.getPhantomReferenceFreedCount();
            G1_PHANTOM_REFERENCE_FREE_COUNT.labels(path).set(phantomReferenceFreedCount);
            double phantomReferencePauseTime = referenceGCSummary.getPhantomReferencePauseTime();
            G1_PHANTOM_REFERENCE_PAUSE_TIME.labels(path).observe(phantomReferencePauseTime);

            int jniWeakReferenceCount = referenceGCSummary.getJniWeakReferenceCount();
            G1_JNI_WEAK_REFERENCE_COUNT.labels(path).set(jniWeakReferenceCount);
            double jniWeakReferencePauseTime = referenceGCSummary.getJniWeakReferencePauseTime();
            G1_JNI_WEAK_REFERENCE_PAUSE_TIME.labels(path).observe(jniWeakReferencePauseTime);
        }

        RegionSummary edenRegion = event.getEdenRegionSummary();
        if (edenRegion != null) {
            if (edenRegion.getBefore() >= 0) {
                G1_EDEN_REGION_BEFORE.labels(path).set(edenRegion.getBefore());
            }
            if (edenRegion.getAfter() >= 0) {
                G1_EDEN_REGION_AFTER.labels(path).set(edenRegion.getAfter());
            }
            if (edenRegion.getAssigned() >= 0) {
                G1_EDEN_REGION_ASSIGN.labels(path).set(edenRegion.getAssigned());
            }
        }
        RegionSummary survivorRegion = event.getSurvivorRegionSummary();
        if (survivorRegion != null) {
            if (survivorRegion.getBefore() >= 0) {
                G1_SURVIVOR_REGION_BEFORE.labels(path).set(survivorRegion.getBefore());
            }
            if (survivorRegion.getAfter() >= 0) {
                G1_SURVIVOR_REGION_AFTER.labels(path).set(survivorRegion.getAfter());
            }
            if (survivorRegion.getAssigned() >= 0) {
                G1_SURVIVOR_REGION_ASSIGN.labels(path).set(survivorRegion.getAssigned());
            }
        }
        RegionSummary oldRegion = event.getOldRegionSummary();
        if (oldRegion != null) {
            if (oldRegion.getBefore() >= 0) {
                G1_OLD_REGION_BEFORE.labels(path).set(oldRegion.getBefore());
            }
            if (oldRegion.getAfter() >= 0) {
                G1_OLD_REGION_AFTER.labels(path).set(oldRegion.getAfter());
            }
            if (oldRegion.getAssigned() >= 0) {
                G1_OLD_REGION_ASSIGN.labels(path).set(oldRegion.getAssigned());
            }
        }
        RegionSummary humongousRegion = event.getHumongousRegionSummary();
        if (humongousRegion != null) {
            if (humongousRegion.getBefore() >= 0) {
                G1_HUMONGOUS_REGION_BEFORE.labels(path).set(humongousRegion.getBefore());
            }
            if (humongousRegion.getAfter() >= 0) {
                G1_HUMONGOUS_REGION_AFTER.labels(path).set(humongousRegion.getAfter());
            }
            if (humongousRegion.getAssigned() >= 0) {
                G1_HUMONGOUS_REGION_ASSIGN.labels(path).set(humongousRegion.getAssigned());
            }
        }
        RegionSummary archiveRegion = event.getArchiveRegionSummary();
        if (archiveRegion != null) {
            if (archiveRegion.getBefore() >= 0) {
                G1_ARCHIVE_REGION_BEFORE.labels(path).set(archiveRegion.getBefore());
            }
            if (archiveRegion.getAfter() >= 0) {
                G1_ARCHIVE_REGION_AFTER.labels(path).set(archiveRegion.getAfter());
            }
            if (archiveRegion.getAssigned() >= 0) {
                G1_ARCHIVE_REGION_ASSIGN.labels(path).set(archiveRegion.getAssigned());
            }
        }
    }

    private String parseGCEventCategory(GenerationalGCEvent event) {

        if (event instanceof GenerationalGCPauseEvent) {

            if (event instanceof FullGC) {
                return "FullGC";
            } else if (event instanceof InitialMark) {
                return "CMSInitialMark";
            } else if (event instanceof CMSRemark) {
                return "CMSRemark";
            } else if (event instanceof DefNew
                    || event instanceof ParNew
                    || event instanceof PSYoungGen
                    || event instanceof YoungGC) {
                return "YoungGC";
            }

        } else if (event instanceof CMSConcurrentEvent) {
            if (event instanceof AbortablePreClean) {
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
        }

        return Collector.sanitizeMetricName(event.getGarbageCollectionType().name());
    }

    private String parseGCEventCategory(G1GCEvent event) {
        if (event instanceof G1GCPauseEvent) {
            if (event instanceof G1Young) {
                if (event instanceof G1YoungInitialMark) {
                    return "G1InitialMark";
                } else if (event instanceof G1Mixed) {
                    return "G1MixedGC";
                }
                return "YoungGC";
            } else if (event instanceof G1Cleanup) {
                return "G1Cleanup";
            } else if (event instanceof G1Remark) {
                return "G1Cleanup";
            } else if (event instanceof G1FullGC) {
                return "FullGC";
            }
        } else if (event instanceof G1GCConcurrentEvent) {
            String category = event.getClass().getSimpleName();
            if (!category.startsWith("G1")) {
                category = "G1" + category;
            }
            return category;
        }
        return Collector.sanitizeMetricName(event.getGarbageCollectionType().name());
    }

    private String parseGCEventCategory(ZGCCycle event) {
        return "zgc";
    }

    // load suitable log parsers for the gc log
    private List<DataSourceParser> loadParsers(File file) {
        try {
            SingleGCLogFile logFile = new SingleGCLogFile(file.toPath());
            Diary diary = logFile.diary();
            if (diary.isG1GC()
                    || diary.isZGC()
                    || diary.isCMS()
                    || diary.isICMS()
                    || diary.isDefNew()
                    || diary.isSerialFull()
                    || diary.isPSOldGen()) {

                List<DataSourceParser> parsers =
                        Arrays.stream(DEFAULT_PARSERS)
                                .map(
                                        parserName -> {
                                            try {
                                                Class<?> clazz =
                                                        forName(
                                                                parserName,
                                                                true,
                                                                Thread.currentThread()
                                                                        .getContextClassLoader());
                                                return Optional.of(
                                                        clazz.getConstructors()[0].newInstance());
                                            } catch (ClassNotFoundException
                                                    | InstantiationException
                                                    | IllegalAccessException
                                                    | InvocationTargetException e) {
                                                return Optional.empty();
                                            }
                                        })
                                .filter(Optional::isPresent)
                                .map(optional -> (DataSourceParser) optional.get())
                                .filter(dataSourceParser -> dataSourceParser.accepts(diary))
                                .collect(Collectors.toList());

                if (!parsers.isEmpty()) {
                    return parsers;
                }
            }
        } catch (Exception ex) {
        }
        throw new UnsupportedOperationException(file.getPath());
    }
}
