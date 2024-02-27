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

import static com.microsoft.gctoolkit.event.GCCause.*;

import com.microsoft.gctoolkit.event.GCCause;
import com.microsoft.gctoolkit.event.g1gc.*;
import com.microsoft.gctoolkit.event.generational.*;
import com.microsoft.gctoolkit.event.zgc.ZGCCycle;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Parsers {
    private static final Logger LOG = LoggerFactory.getLogger(Parsers.class);

    private Parsers() {}

    public static String parseGenerationalGCEventCategory(GenerationalGCEvent event) {
        if (isCMSGCEvent(event)) {
            return parseCMSGCEventCategory(event);
        }
        return parseOtherGenerationalGCEventCategory(event);
    }

    private static boolean isCMSGCEvent(GenerationalGCEvent event) {
        return event instanceof CMSPhase
                || event instanceof ConcurrentModeFailure
                || event instanceof ConcurrentModeInterrupted;
    }

    private static String parseCMSGCEventCategory(GenerationalGCEvent event) {
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

    private static String parseOtherGenerationalGCEventCategory(GenerationalGCEvent event) {
        if (event instanceof GenerationalGCPauseEvent) {
            if (event instanceof FullGC) {
                if (event instanceof PSFullGC) {
                    return "PSFullGC";
                } else if (event instanceof SystemGC) {
                    return "SystemGC";
                }
                return "FullGC";
            } else if (event instanceof ParNew) {
                if (event instanceof ParNewPromotionFailed) {
                    return "ParNewPromotionFailed";
                }
                return "ParNew";
            } else if (event instanceof PSYoungGen) {
                return "PSYoungGC";
            } else if (event instanceof DefNew) {
                return "DefNew";
            } else if (event instanceof YoungGC) {
                return "YoungGC";
            }
        }
        return "Unknown";
    }

    public static String parseG1GCEventCategory(G1GCEvent event) {
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

    public static String parseZGCEventCategory(ZGCCycle event) {
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
