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
package prometheus.exporter.jgc.tool;

import io.prometheus.client.Counter;
import io.prometheus.client.Gauge;
import io.prometheus.client.Summary;

public class Metrics {
    private Metrics() {}

    public static final Gauge STARTUP =
            Gauge.build().name("jgc_startup").help("Timestamp of exporter startup").register();

    public static final Counter GC_LOG_LINE =
            Counter.build()
                    .labelNames("path")
                    .name("jgc_log_line")
                    .help("Number of process log lines")
                    .register();
    public static final Summary GC_EVENT_DURATION =
            Summary.build()
                    .labelNames("path", "category")
                    .name("jgc_event_duration")
                    .help("Duration(ms) of gc event")
                    .register();

    public static final Summary USER_CPU_TIME =
            Summary.build().labelNames("path").name("jgc_user_cpu_time").help("help").register();

    public static final Summary SYS_CPU_TIME =
            Summary.build().labelNames("path").name("jgc_sys_cpu_time").help("help").register();

    public static final Summary REAL_CPU_TIME =
            Summary.build().labelNames("path").name("jgc_real_cpu_time").help("help").register();

    public static final Gauge GENERATIONAL_YOUNG_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_young_occupancy_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_YOUNG_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_young_size_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_YOUNG_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_young_occupancy_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_YOUNG_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_young_size_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_TENURED_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_tenured_occupancy_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_TENURED_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_tenured_size_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_TENURED_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_tenured_occupancy_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_TENURED_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_tenured_size_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_HEAP_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_heap_occupancy_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_HEAP_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_heap_size_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_HEAP_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_heap_occupancy_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_HEAP_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_heap_size_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_CLASSSPACE_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_classspace_occupancy_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_CLASSSPACE_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_classspace_size_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_CLASSSPACE_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_classspace_occupancy_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_CLASSSPACE_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_classspace_size_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_NONCLASSSPACE_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_nonclassspace_occupancy_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_NONCLASSSPACE_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_nonclassspace_size_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_NONCLASSSPACE_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_nonclassspace_occupancy_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_NONCLASSSPACE_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_nonclassspace_size_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_METASPACE_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_metaspace_occupancy_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_METASPACE_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_metaspace_size_before_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_METASPACE_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_metaspace_occupancy_after_collection")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_METASPACE_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_metaspace_size_after_collection")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_CLASS_UNLOADING_PROCESS_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_class_unloading_process_time")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_SYMBOL_TABLE_PROCESS_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_symbol_table_process_time")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_STRING_TABLE_PROCESS_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_string_table_process_time")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_SYMBOL_AND_STRING_TABLE_PROCESS_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_symbol_and_string_table_process_time")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_SOFT_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_soft_reference_count")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_SOFT_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_soft_reference_pause_time")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_WEAK_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_weak_reference_count")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_WEAK_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_weak_reference_pause_time")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_FINAL_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_final_reference_count")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_FINAL_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_final_reference_pause_time")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_PHANTOM_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_phantom_reference_count")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_PHANTOM_REFERENCE_FREE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_phantom_reference_free_count")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_PHANTOM_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_phantom_reference_pause_time")
                    .help("help")
                    .register();

    public static final Gauge GENERATIONAL_JNI_WEAK_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_generational_jni_weak_reference_count")
                    .help("help")
                    .register();

    public static final Summary GENERATIONAL_JNI_WEAK_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_generational_jni_weak_reference_pause_time")
                    .help("help")
                    .register();

    public static final Gauge G1_EDEN_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_eden_occupancy_after_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_EDEN_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_eden_heap_occupancy_before_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_EDEN_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_eden_size_after_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_EDEN_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_eden_size_before_collection_bytes")
                    .help("help")
                    .register();
    public static final Gauge G1_SURVIVOR_HEAP_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_survivor_heap_occupancy_after_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_SURVIVOR_HEAP_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_survivor_heap_occupancy_before_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_SURVIVOR_SIZE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_survivor_size_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_META_SPACE_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_meta_space_occupancy_after_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_META_SPACE_OCCUPANCY_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_meta_space_heap_occupancy_before_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_META_SPACE_SIZE_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_meta_space_size_after_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_META_SPACE_SIZE_BEFORE_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_meta_space_size_before_collection_bytes")
                    .help("help")
                    .register();

    public static final Gauge G1_SOFT_REFERENCE_COUNT =
            Gauge.build().labelNames("path").name("jgc_g1_soft_references").help("help").register();

    public static final Summary G1_SOFT_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_g1_soft_reference_pause_duration_seconds")
                    .help("help")
                    .register();

    public static final Gauge G1_WEAK_REFERENCE_COUNT =
            Gauge.build().labelNames("path").name("jgc_g1_weak_references").help("help").register();

    public static final Summary G1_WEAK_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_g1_weak_reference_pause_duration_seconds")
                    .help("help")
                    .register();

    public static final Gauge G1_FINAL_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_final_references")
                    .help("help")
                    .register();

    public static final Summary G1_FINAL_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_g1_final_reference_pause_duration_seconds")
                    .help("help")
                    .register();

    public static final Gauge G1_PHANTOM_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_phantom_references")
                    .help("help")
                    .register();

    public static final Gauge G1_PHANTOM_REFERENCE_FREE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_phantom_free_references")
                    .help("help")
                    .register();

    public static final Summary G1_PHANTOM_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_g1_phantom_reference_pause_duration_seconds")
                    .help("help")
                    .register();

    public static final Gauge G1_JNI_WEAK_REFERENCE_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_jni_weak_references")
                    .help("help")
                    .register();

    public static final Summary G1_JNI_WEAK_REFERENCE_PAUSE_TIME =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_g1_jni_weak_reference_pause_duration_seconds")
                    .help("help")
                    .register();

    public static final Gauge G1_EDEN_REGION_BEFORE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_eden_region_before")
                    .help("help")
                    .register();

    public static final Gauge G1_EDEN_REGION_AFTER =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_eden_region_after")
                    .help("help")
                    .register();

    public static final Gauge G1_EDEN_REGION_ASSIGN =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_eden_region_assign")
                    .help("help")
                    .register();

    public static final Gauge G1_SURVIVOR_REGION_BEFORE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_survivor_region_before")
                    .help("help")
                    .register();

    public static final Gauge G1_SURVIVOR_REGION_AFTER =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_survivor_region_after")
                    .help("help")
                    .register();

    public static final Gauge G1_SURVIVOR_REGION_ASSIGN =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_survivor_region_assign")
                    .help("help")
                    .register();

    public static final Gauge G1_OLD_REGION_BEFORE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_old_region_before")
                    .help("help")
                    .register();

    public static final Gauge G1_OLD_REGION_AFTER =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_old_region_after")
                    .help("help")
                    .register();

    public static final Gauge G1_OLD_REGION_ASSIGN =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_old_region_assign")
                    .help("help")
                    .register();

    public static final Gauge G1_HUMONGOUS_REGION_BEFORE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_humongous_region_before")
                    .help("help")
                    .register();

    public static final Gauge G1_HUMONGOUS_REGION_AFTER =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_humongous_region_after")
                    .help("help")
                    .register();

    public static final Gauge G1_HUMONGOUS_REGION_ASSIGN =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_humongous_region_assign")
                    .help("help")
                    .register();

    public static final Gauge G1_ARCHIVE_REGION_BEFORE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_archive_region_before")
                    .help("help")
                    .register();

    public static final Gauge G1_ARCHIVE_REGION_AFTER =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_archive_region_after")
                    .help("help")
                    .register();

    public static final Gauge G1_ARCHIVE_REGION_ASSIGN =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_g1_archive_region_assign")
                    .help("help")
                    .register();
    public static final Summary ZGC_PAUSE_MARK_START_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_pause_mark_start_duration")
                    .help("help")
                    .register();
    public static final Summary ZGC_CONCURRENT_MARK_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_concurrent_mark_duration")
                    .help("help")
                    .register();

    public static final Summary ZGC_CONCURRENT_MARK_FREE_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_concurrent_mark_free_duration")
                    .help("help")
                    .register();

    public static final Summary ZGC_PAUSE_MARK_END_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_pause_mark_end_duration")
                    .help("help")
                    .register();

    public static final Summary ZGC_PROCESS_NON_STRONG_REFERENCES_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_process_non_strong_references_duration")
                    .help("help")
                    .register();

    public static final Summary ZGC_CONCURRENT_RESET_RELOCATIONSET_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_concurrent_reset_relocationset_duration")
                    .help("help")
                    .register();

    public static final Summary ZGC_CONCURRENT_SELECT_RELOCATIONSET_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_concurrent_select_relocationset_duration")
                    .help("help")
                    .register();

    public static final Summary ZGC_PAUSE_RELOCATE_START_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_pause_relocate_start_duration")
                    .help("help")
                    .register();

    public static final Summary ZGC_CONCURRENT_RELOCATE_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_zgc_concurrent_relocate_duration")
                    .help("help")
                    .register();

    public static final Gauge ZGC_LOAD_1m =
            Gauge.build().labelNames("path").name("jgc_zgc_load_1m").help("help").register();

    public static final Gauge ZGC_LOAD_5m =
            Gauge.build().labelNames("path").name("jgc_zgc_load_5m").help("help").register();

    public static final Gauge ZGC_LOAD_15m =
            Gauge.build().labelNames("path").name("jgc_zgc_load_15m").help("help").register();

    public static final Gauge ZGC_MMU_2MS =
            Gauge.build().labelNames("path").name("jgc_zgc_mmu_2ms").help("help").register();

    public static final Gauge ZGC_MMU_5MS =
            Gauge.build().labelNames("path").name("jgc_zgc_mmu_5ms").help("help").register();

    public static final Gauge ZGC_MMU_10MS =
            Gauge.build().labelNames("path").name("jgc_zgc_mmu_10ms").help("help").register();

    public static final Gauge ZGC_MMU_20MS =
            Gauge.build().labelNames("path").name("jgc_zgc_mmu_20ms").help("help").register();

    public static final Gauge ZGC_MMU_50MS =
            Gauge.build().labelNames("path").name("jgc_zgc_mmu_50ms").help("help").register();

    public static final Gauge ZGC_MMU_100MS =
            Gauge.build().labelNames("path").name("jgc_zgc_mmu_100ms").help("help").register();

    public static final Gauge ZGC_MARK_START_USED =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_mark_start_used")
                    .help("help")
                    .register();

    public static final Gauge ZGC_MARK_START_FREE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_mark_start_free")
                    .help("help")
                    .register();

    public static final Gauge ZGC_MARK_END_USED =
            Gauge.build().labelNames("path").name("jgc_zgc_mark_end_used").help("help").register();

    public static final Gauge ZGC_MARK_END_FREE =
            Gauge.build().labelNames("path").name("jgc_zgc_mark_end_free").help("help").register();

    public static final Gauge ZGC_RELOCATE_START_USED =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_relocate_start_used")
                    .help("help")
                    .register();

    public static final Gauge ZGC_RELOCATE_START_FREE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_relocate_start_free")
                    .help("help")
                    .register();

    public static final Gauge ZGC_RELOCATE_END_USED =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_relocate_end_used")
                    .help("help")
                    .register();

    public static final Gauge ZGC_RELOCATE_END_FREE =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_relocate_end_free")
                    .help("help")
                    .register();

    public static final Gauge ZGC_LIVE_MARK_END =
            Gauge.build().labelNames("path").name("jgc_zgc_live_mark_end").help("help").register();

    public static final Gauge ZGC_LIVE_RECLAIM_START =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_live_reclaim_start")
                    .help("help")
                    .register();

    public static final Gauge ZGC_LIVE_RECLAIM_END =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_live_reclaim_end")
                    .help("help")
                    .register();

    public static final Gauge ZGC_ALLOCATED_MARK_END =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_allocated_mark_end")
                    .help("help")
                    .register();

    public static final Gauge ZGC_ALLOCATED_RECLAIM_START =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_allocated_reclaim_start")
                    .help("help")
                    .register();

    public static final Gauge ZGC_ALLOCATED_RECLAIM_END =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_allocated_reclaim_end")
                    .help("help")
                    .register();

    public static final Gauge ZGC_GARBAGE_MARK_END =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_garbage_mark_end")
                    .help("help")
                    .register();

    public static final Gauge ZGC_GARBAGE_RECLAIM_START =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_garbage_reclaim_start")
                    .help("help")
                    .register();

    public static final Gauge ZGC_GARBAGE_RECLAIM_END =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_garbage_reclaim_end")
                    .help("help")
                    .register();

    public static final Gauge ZGC_RECLAIMED_RECLAIM_START =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_reclaimed_reclaim_start")
                    .help("help")
                    .register();

    public static final Gauge ZGC_RECLAIMED_RECLAIM_END =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_reclaimed_reclaim_end")
                    .help("help")
                    .register();

    public static final Gauge ZGC_MEMORY_RECLAIM_START =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_memory_reclaim_start")
                    .help("help")
                    .register();

    public static final Gauge ZGC_MEMORY_RECLAIM_END =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_memory_reclaim_end")
                    .help("help")
                    .register();

    public static final Gauge ZGC_METASPACE_USED =
            Gauge.build().labelNames("path").name("jgc_zgc_metaspace_used").help("help").register();

    public static final Gauge ZGC_METASPACE_COMMITTED =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_metaspace_committed")
                    .help("help")
                    .register();

    public static final Gauge ZGC_METASPACE_RESERVED =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_zgc_metaspace_reserved")
                    .help("help")
                    .register();

    public static final Gauge SAFEPOINT_TOTAL_NUMBER_OF_APPLICATION_THREADS =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_safepoint_total_number_of_application_threads")
                    .help("help")
                    .register();

    public static final Gauge SAFEPOINT_INITIALLY_RUNNING =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_safepoint_initially_running")
                    .help("help")
                    .register();

    public static final Gauge SAFEPOINT_WAITING_TO_BLOCK =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_safepoint_waiting_to_block")
                    .help("help")
                    .register();

    public static final Summary SAFEPOINT_SPIN_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_safepoint_spin_duration")
                    .help("help")
                    .register();

    public static final Summary SAFEPOINT_BLOCK_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_safepoint_block_duration")
                    .help("help")
                    .register();

    public static final Summary SAFEPOINT_SYNC_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_safepoint_sync_duration")
                    .help("help")
                    .register();

    public static final Summary SAFEPOINT_CLEANUP_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_safepoint_cleanup_duration")
                    .help("help")
                    .register();

    public static final Summary SAFEPOINT_VMOP_DURATION =
            Summary.build()
                    .labelNames("path")
                    .name("jgc_safepoint_vmop_duration")
                    .help("help")
                    .register();

    public static final Gauge SAFEPOINT_PAGE_TRAP_COUNT =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_safepoint_page_trap_count")
                    .help("help")
                    .register();

    public static final Gauge SURVIVOR_DESIRED_OCCUPANCY_AFTER_COLLECTION =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_survivor_desired_occupancy_after_collection")
                    .help("help")
                    .register();

    public static final Gauge SURVIVOR_CALCULATED_TENURING_THRESHOLD =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_survivor_calculated_tenuring_threshold")
                    .help("help")
                    .register();

    public static final Gauge SURVIVOR_MAX_TENURING_THRESHOLD =
            Gauge.build()
                    .labelNames("path")
                    .name("jgc_survivor_max_tenuring_threshold")
                    .help("help")
                    .register();
}
