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

import static com.microsoft.gctoolkit.event.GCCause.*;
import static prometheus.exporter.jgc.metric.MetricRegistry.*;

import com.microsoft.gctoolkit.event.GCCause;
import com.microsoft.gctoolkit.event.jvm.JVMEvent;
import com.microsoft.gctoolkit.event.zgc.*;
import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.message.ChannelName;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.parser.ZGCParser;
import java.io.File;
import java.util.Collections;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ZGCEventHandler extends AbstractJVMEventHandler {
    private static final Logger LOG = LoggerFactory.getLogger(ZGCEventHandler.class);

    public ZGCEventHandler(File file, Diary diary) {
        super(file, diary);
    }

    @Override
    protected List<DataSourceParser> loadParsers() {
        return Collections.singletonList(new ZGCParser());
    }

    @Override
    public void publish(ChannelName channel, JVMEvent event) {
        if (event instanceof ZGCCycle) {
            recordZGCEvent((ZGCCycle) event);
        } else {
            LOG.warn("{} published an unsupported event: {}", channel, event);
        }
    }

    void recordZGCEvent(ZGCCycle event) {
        final String category = parseCategory(event);
        LOG.debug("Collect ZGCCycle: {} ", category);

        double duration = 0;
        double pauseDuration = 0;
        double pauseMarkStartDuration = event.getPauseMarkStartDuration() / 1000;
        duration += pauseMarkStartDuration;
        pauseDuration += pauseMarkStartDuration;
        ZGC_PAUSE_MARK_START_DURATION.attach(this, path, host).observe(pauseMarkStartDuration);
        double concurrentMarkDuration = event.getConcurrentMarkDuration() / 1000;
        duration += concurrentMarkDuration;
        ZGC_CONCURRENT_MARK_DURATION.attach(this, path, host).observe(concurrentMarkDuration);
        double concurrentMarkFreeDuration = event.getConcurrentMarkFreeDuration() / 1000;
        duration += concurrentMarkFreeDuration;
        ZGC_CONCURRENT_MARK_FREE_DURATION
                .attach(this, path, host)
                .observe(concurrentMarkFreeDuration);
        double pauseMarkEndDuration = event.getPauseMarkEndDuration() / 1000;
        duration += pauseMarkEndDuration;
        pauseDuration += pauseMarkEndDuration;
        ZGC_PAUSE_MARK_END_DURATION.attach(this, path, host).observe(pauseMarkEndDuration);
        double concurrentProcessNonStrongReferencesDuration =
                event.getConcurrentProcessNonStrongReferencesDuration() / 1000;
        duration += concurrentProcessNonStrongReferencesDuration;
        ZGC_PROCESS_NON_STRONG_REFERENCES_DURATION
                .attach(this, path, host)
                .observe(concurrentProcessNonStrongReferencesDuration);
        double concurrentResetRelocationSetDuration =
                event.getConcurrentResetRelocationSetDuration() / 1000;
        duration += concurrentResetRelocationSetDuration;
        ZGC_CONCURRENT_RESET_RELOCATIONSET_DURATION
                .attach(this, path, host)
                .observe(concurrentResetRelocationSetDuration);
        double concurrentSelectRelocationSetDuration =
                event.getConcurrentSelectRelocationSetDuration() / 1000;
        duration += concurrentSelectRelocationSetDuration;
        ZGC_CONCURRENT_SELECT_RELOCATIONSET_DURATION
                .attach(this, path, host)
                .observe(concurrentSelectRelocationSetDuration);
        double pauseRelocateStartDuration = event.getPauseRelocateStartDuration() / 1000;
        duration += pauseRelocateStartDuration;
        pauseDuration += pauseRelocateStartDuration;
        ZGC_PAUSE_RELOCATE_START_DURATION
                .attach(this, path, host)
                .observe(pauseRelocateStartDuration);
        double concurrentRelocateDuration = event.getConcurrentRelocateDuration() / 1000;
        duration += concurrentRelocateDuration;
        ZGC_CONCURRENT_RELOCATE_DURATION
                .attach(this, path, host)
                .observe(concurrentRelocateDuration);

        GC_EVENT_DURATION.attach(this, path, host, category).observe(duration);
        GC_EVENT_LAST_MINUTE_DURATION.attach(this, path, host).observe(duration);
        GC_EVENT_PAUSE_DURATION.attach(this, path, host, category).observe(pauseDuration);
        ;

        double load1m = event.getLoadAverageAt(1);
        double load5m = event.getLoadAverageAt(5);
        double load15m = event.getLoadAverageAt(15);
        ZGC_LOAD_1m.attach(this, path, host).set(load1m);
        ZGC_LOAD_5m.attach(this, path, host).set(load5m);
        ZGC_LOAD_15m.attach(this, path, host).set(load15m);

        double mmu_2ms = event.getMMU(2) / 100d;
        double mmu_5ms = event.getMMU(5) / 100d;
        double mmu_10ms = event.getMMU(10) / 100d;
        double mmu_20ms = event.getMMU(20) / 100d;
        double mmu_50ms = event.getMMU(50) / 100d;
        double mmu_100ms = event.getMMU(100) / 100d;
        ZGC_MMU_2MS.attach(this, path, host).set(mmu_2ms);
        ZGC_MMU_5MS.attach(this, path, host).set(mmu_5ms);
        ZGC_MMU_10MS.attach(this, path, host).set(mmu_10ms);
        ZGC_MMU_20MS.attach(this, path, host).set(mmu_20ms);
        ZGC_MMU_50MS.attach(this, path, host).set(mmu_50ms);
        ZGC_MMU_100MS.attach(this, path, host).set(mmu_100ms);

        ZGCMemoryPoolSummary markStart = event.getMarkStart();
        if (markStart != null) {
            ZGC_MARK_START_USED.attach(this, path, host).set(markStart.getUsed() * 1024);
            ZGC_MARK_START_FREE.attach(this, path, host).set(markStart.getFree() * 1024);
            HEAP_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(markStart.getUsed() * 1024);
            HEAP_SIZE_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set((markStart.getUsed() + markStart.getFree()) * 1024);
        }
        ZGCMemoryPoolSummary markEnd = event.getMarkEnd();
        if (markEnd != null) {
            ZGC_MARK_END_USED.attach(this, path, host).set(markEnd.getUsed() * 1024);
            ZGC_MARK_END_FREE.attach(this, path, host).set(markEnd.getFree() * 1024);
        }
        ZGCMemoryPoolSummary relocateStart = event.getRelocateStart();
        if (relocateStart != null) {
            ZGC_RELOCATE_START_USED.attach(this, path, host).set(relocateStart.getUsed() * 1024);
            ZGC_RELOCATE_START_FREE.attach(this, path, host).set(relocateStart.getFree() * 1024);
        }
        ZGCMemoryPoolSummary relocateEnd = event.getRelocateStart();
        if (relocateEnd != null) {
            ZGC_RELOCATE_END_USED.attach(this, path, host).set(relocateEnd.getUsed() * 1024);
            ZGC_RELOCATE_END_FREE.attach(this, path, host).set(relocateEnd.getFree() * 1024);
            HEAP_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(relocateEnd.getUsed() * 1024);
            HEAP_SIZE_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set((relocateEnd.getUsed() + relocateEnd.getFree()) * 1024);
        }
        OccupancySummary live = event.getLive();
        if (live != null) {
            ZGC_LIVE_MARK_END.attach(this, path, host).set(live.getMarkEnd());
            ZGC_LIVE_RECLAIM_START.attach(this, path, host).set(live.getReclaimStart() * 1024);
            ZGC_LIVE_RECLAIM_END.attach(this, path, host).set(live.getReclaimEnd() * 1024);
        }
        OccupancySummary allocated = event.getAllocated();
        if (allocated != null) {
            ZGC_ALLOCATED_MARK_END.attach(this, path, host).set(allocated.getMarkEnd() * 1024);
            ZGC_ALLOCATED_RECLAIM_START
                    .attach(this, path, host)
                    .set(allocated.getReclaimStart() * 1024);
            ZGC_ALLOCATED_RECLAIM_END
                    .attach(this, path, host)
                    .set(allocated.getReclaimEnd() * 1024);
        }
        OccupancySummary garbage = event.getGarbage();
        if (garbage != null) {
            ZGC_GARBAGE_MARK_END.attach(this, path, host).set(garbage.getMarkEnd() * 1024);
            ZGC_GARBAGE_RECLAIM_START
                    .attach(this, path, host)
                    .set(garbage.getReclaimStart() * 1024);
            ZGC_GARBAGE_RECLAIM_END.attach(this, path, host).set(garbage.getReclaimEnd() * 1024);
        }
        ReclaimSummary reclaimed = event.getReclaimed();
        if (reclaimed != null) {
            ZGC_RECLAIMED_RECLAIM_START
                    .attach(this, path, host)
                    .set(reclaimed.getReclaimStart() * 1024);
            ZGC_RECLAIMED_RECLAIM_END
                    .attach(this, path, host)
                    .set(reclaimed.getReclaimEnd() * 1024);
        }
        ReclaimSummary memorySummary = event.getMemorySummary();
        if (memorySummary != null) {
            ZGC_MEMORY_RECLAIM_START
                    .attach(this, path, host)
                    .set(memorySummary.getReclaimStart() * 1024);
            ZGC_MEMORY_RECLAIM_END
                    .attach(this, path, host)
                    .set(memorySummary.getReclaimEnd() * 1024);
        }
        ZGCMetaspaceSummary metaspace = event.getMetaspace();
        if (metaspace != null) {
            ZGC_METASPACE_USED.attach(this, path, host).set(metaspace.getUsed() * 1024);
            ZGC_METASPACE_COMMITTED.attach(this, path, host).set(metaspace.getCommitted() * 1024);
            ZGC_METASPACE_RESERVED.attach(this, path, host).set(metaspace.getReserved() * 1024);
        }
    }

    private String parseCategory(ZGCCycle event) {
        GCCause cause = event.getGCCause();
        if (cause == TIMER) {
            return "ZGCTimer";
        } else if (cause == WARMUP) {
            return "ZGCWarmup";
        } else if (cause == ALLOC_RATE) {
            return "ZGCAllocRate";
        } else if (cause == ALLOC_STALL) {
            return "ZGCAllocStall";
        } else if (cause == PROACTIVE) {
            return "ZGCProactive";
        } else if (cause == METADATA_GENERATION_THRESHOLD) {
            return "ZGCMetadataGCThreshold";
        } else if (cause == JAVA_LANG_SYSTEM) {
            return "ZGCSystemGc";
        }
        return "ZGCUnknown";
    }
}
