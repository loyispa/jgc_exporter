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
package prometheus.exporter.jgc.collector.metric;

import io.prometheus.client.*;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

public class CollectorProxyRegistry extends CollectorRegistry {
    public static final CollectorProxyRegistry SINGLETON = new CollectorProxyRegistry();

    public static final CollectorProxy<Gauge.Child, Gauge> EXPORTER_STARTUP_SECONDS =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .name("jgc_startup_timestamp_seconds")
                                    .help("Timestamp of exporter startup")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> EXPORTER_VERSION_INFO =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("version")
                                    .name("jgc_exporter_version_info")
                                    .help("The version of jgc_exporter")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> GC_COLLECT_FILES =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_collect_files")
                                    .help("jgc exporter collect file list")
                                    .create());
    public static final CollectorProxy<Counter.Child, Counter> GC_LOG_LINES =
            CollectorProxy.of(
                    () ->
                            Counter.build()
                                    .labelNames("path")
                                    .name("jgc_log_lines")
                                    .help("Number of process log lines")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> GC_EVENT_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .name("jgc_event_duration_seconds")
                                    .help("Duration of gc event")
                                    .labelNames("path", "category")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> GC_EVENT_LAST_MINUTE_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .ageBuckets(6)
                                    .maxAgeSeconds(60)
                                    .quantile(0, 0.05)
                                    .quantile(0.5, 0.05)
                                    .quantile(0.75, 0.05)
                                    .quantile(1.0, 0.05)
                                    .labelNames("path")
                                    .name("jgc_event_last_minute_duration_seconds")
                                    .help("Last minute duration of gc event")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> GC_EVENT_PAUSE_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path", "category")
                                    .name("jgc_event_pause_duration_seconds")
                                    .help("Duration of gc pause event")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> HEAP_OCCUPANCY_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_heap_occupancy_before_collection_bytes")
                                    .help("heap occupancy before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> HEAP_SIZE_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_heap_size_before_collection_bytes")
                                    .help("heap size before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> HEAP_OCCUPANCY_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_heap_occupancy_after_collection_bytes")
                                    .help("heap occupancy after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> HEAP_SIZE_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_heap_size_after_collection_bytes")
                                    .help("heap size after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> YOUNG_OCCUPANCY_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_young_occupancy_before_collection_bytes")
                                    .help("young generation occupancy before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> YOUNG_SIZE_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_generational_young_size_before_collection_bytes")
                                    .help("young generation size before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> YOUNG_OCCUPANCY_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_generational_young_occupancy_after_collection_bytes")
                                    .help("young generation occupancy after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> YOUNG_SIZE_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_generational_young_size_after_collection_bytes")
                                    .help("young generation size after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> OLD_OCCUPANCY_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_old_occupancy_before_collection_bytes")
                                    .help("old generation occupancy before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> OLD_SIZE_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_old_size_before_collection_bytes")
                                    .help("old generation size before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> OLD_OCCUPANCY_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_old_occupancy_after_collection_bytes")
                                    .help("old generation occupancy after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> OLD_SIZE_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_old_size_after_collection_bytes")
                                    .help("old generation size after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> METASPACE_OCCUPANCY_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_metaspace_occupancy_before_collection_bytes")
                                    .help("metaspace occupancy before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> METASPACE_SIZE_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_metaspace_size_before_collection_bytes")
                                    .help("metaspace size before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> METASPACE_OCCUPANCY_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_metaspace_occupancy_after_collection_bytes")
                                    .help("metaspace occupancy after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> METASPACE_SIZE_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_metaspace_size_after_collection_bytes")
                                    .help("metaspace size after collection")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> CMS_CLASS_UNLOADING_PROCESS_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_cms_class_unloading_process_duration_seconds")
                                    .help("class unloading process time")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> CMS_SYMBOL_TABLE_PROCESS_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_cms_symbol_table_process_duration_seconds")
                                    .help("symbol table process time")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> CMS_STRING_TABLE_PROCESS_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_cms_string_table_process_duration_seconds")
                                    .help("string table process duration")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary>
            CMS_SYMBOL_AND_STRING_TABLE_PROCESS_TIME =
                    CollectorProxy.of(
                            () ->
                                    Summary.build()
                                            .labelNames("path")
                                            .name("jgc_cms_symbol_and_string_table_process_seconds")
                                            .help("symbol and string table process duration")
                                            .create());

    public static final CollectorProxy<Gauge.Child, Gauge> SOFT_REFERENCE_COUNT =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_soft_references")
                                    .help("amount of soft references")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> SOFT_REFERENCE_PAUSE_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_soft_reference_pause_duration_seconds")
                                    .help("soft reference pause duration")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> WEAK_REFERENCE_COUNT =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_weak_references")
                                    .help("amount of weak references")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> WEAK_REFERENCE_PAUSE_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_weak_reference_pause_seconds")
                                    .help("weak reference pause duration")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> FINAL_REFERENCE_COUNT =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_final_references")
                                    .help("amount of final references")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> FINAL_REFERENCE_PAUSE_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_final_reference_pause_duration_seconds")
                                    .help("final reference pause duration")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> PHANTOM_REFERENCE_COUNT =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_phantom_references")
                                    .help("amount of phantom references")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> PHANTOM_REFERENCE_FREE_COUNT =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_free_phantom_references")
                                    .help("amount of free phantom references")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> PHANTOM_REFERENCE_PAUSE_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_phantom_reference_pause_duration_seconds")
                                    .help("phantom reference pause duration")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> JNI_WEAK_REFERENCE_COUNT =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_jni_weak_references")
                                    .help("amount of jni weak references")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> JNI_WEAK_REFERENCE_PAUSE_TIME =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_jni_weak_reference_pause_duration_seconds")
                                    .help("jni weak reference pause duration")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_EDEN_OCCUPANCY_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_eden_occupancy_after_collection_bytes")
                                    .help("eden occupancy bytes after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_EDEN_OCCUPANCY_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_eden_heap_occupancy_before_collection_bytes")
                                    .help("eden heap occupancy bytes before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_EDEN_SIZE_AFTER_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_eden_size_after_collection_bytes")
                                    .help("eden size after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_EDEN_SIZE_BEFORE_COLLECTION =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_eden_size_before_collection_bytes")
                                    .help("eden size before collection")
                                    .create());
    public static final CollectorProxy<Gauge.Child, Gauge>
            G1_SURVIVOR_HEAP_OCCUPANCY_AFTER_COLLECTION =
                    CollectorProxy.of(
                            () ->
                                    Gauge.build()
                                            .labelNames("path")
                                            .name(
                                                    "jgc_g1_survivor_heap_occupancy_after_collection_bytes")
                                            .help("survivor heap occupancy bytes after collection")
                                            .create());

    public static final CollectorProxy<Gauge.Child, Gauge>
            G1_SURVIVOR_HEAP_OCCUPANCY_BEFORE_COLLECTION =
                    CollectorProxy.of(
                            () ->
                                    Gauge.build()
                                            .labelNames("path")
                                            .name(
                                                    "jgc_g1_survivor_heap_occupancy_before_collection_bytes")
                                            .help("survivor heap occupancy bytes before collection")
                                            .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_SURVIVOR_SIZE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_survivor_size_bytes")
                                    .help("survivor size")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_EDEN_REGION_BEFORE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_eden_before_collection_regions")
                                    .help("amount of g1 eden region before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_EDEN_REGION_AFTER =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_eden_after_collection_regions")
                                    .help("amount of g1 eden region after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_EDEN_REGION_ASSIGN =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_eden_assign_regions")
                                    .help("amount of g1 eden assign regions")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_SURVIVOR_REGION_BEFORE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_survivor_before_collection_regions")
                                    .help("amount of g1 survivor region before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_SURVIVOR_REGION_AFTER =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_survivor_after_collection_regions")
                                    .help("amount of g1 survivor region after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_SURVIVOR_REGION_ASSIGN =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_survivor_assign_regions")
                                    .help("amount of g1 survivor assign regions")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_OLD_REGION_BEFORE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_old_before_collection_regions")
                                    .help("amount of g1 old regions before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_OLD_REGION_AFTER =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_old_after_collection_regions")
                                    .help("amount of g1 old regions after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_OLD_REGION_ASSIGN =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_old_assign_regions")
                                    .help("amount of g1 old assign regions")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_HUMONGOUS_REGION_BEFORE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_humongous_before_collection_regions")
                                    .help("amount of g1 humongous regions before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_HUMONGOUS_REGION_AFTER =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_humongous_after_collection_regions")
                                    .help("amount of g1 humongous regions after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_HUMONGOUS_REGION_ASSIGN =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_humongous_assign_regions")
                                    .help("amount of g1 humongous assign regions")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_ARCHIVE_REGION_BEFORE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_archive_before_collection_regions")
                                    .help("amount of g1 archive regions before collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_ARCHIVE_REGION_AFTER =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_archive_after_collection_regions")
                                    .help("amount of g1 archive regions after collection")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> G1_ARCHIVE_REGION_ASSIGN =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_g1_archive_assign_regions")
                                    .help("amount of g1 archive assign regions")
                                    .create());
    public static final CollectorProxy<Summary.Child, Summary> ZGC_PAUSE_MARK_START_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_pause_mark_start_duration_seconds")
                                    .help("zgc pause mark start duration")
                                    .create());
    public static final CollectorProxy<Summary.Child, Summary> ZGC_CONCURRENT_MARK_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_concurrent_mark_duration_seconds")
                                    .help("zgc concurrent mark duration")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> ZGC_CONCURRENT_MARK_FREE_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_concurrent_mark_free_duration_seconds")
                                    .help("zgc concurrent mark free duration")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> ZGC_PAUSE_MARK_END_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_pause_mark_end_duration_seconds")
                                    .help("zgc concurrent mark end duration")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary>
            ZGC_PROCESS_NON_STRONG_REFERENCES_DURATION =
                    CollectorProxy.of(
                            () ->
                                    Summary.build()
                                            .labelNames("path")
                                            .name(
                                                    "jgc_zgc_process_non_strong_references_duration_seconds")
                                            .help("zgc process non-strong references duration")
                                            .create());

    public static final CollectorProxy<Summary.Child, Summary>
            ZGC_CONCURRENT_RESET_RELOCATIONSET_DURATION =
                    CollectorProxy.of(
                            () ->
                                    Summary.build()
                                            .labelNames("path")
                                            .name(
                                                    "jgc_zgc_concurrent_reset_relocationset_duration_seconds")
                                            .help("zgc concurrent reset relocationset duration")
                                            .create());

    public static final CollectorProxy<Summary.Child, Summary>
            ZGC_CONCURRENT_SELECT_RELOCATIONSET_DURATION =
                    CollectorProxy.of(
                            () ->
                                    Summary.build()
                                            .labelNames("path")
                                            .name(
                                                    "jgc_zgc_concurrent_select_relocationset_duration_seconds")
                                            .help("zgc concurrent select relocationset duration")
                                            .create());

    public static final CollectorProxy<Summary.Child, Summary> ZGC_PAUSE_RELOCATE_START_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_pause_relocate_start_duration_seconds")
                                    .help("zgc pause relocate start duration")
                                    .create());

    public static final CollectorProxy<Summary.Child, Summary> ZGC_CONCURRENT_RELOCATE_DURATION =
            CollectorProxy.of(
                    () ->
                            Summary.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_concurrent_relocate_duration_seconds")
                                    .help("zgc concurrent relocate duration")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_LOAD_1m =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_1m_cpu_load")
                                    .help("zgc latest 1 minute cpu load average")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_LOAD_5m =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_5m_cpu_load")
                                    .help("zgc latest 5 minute cpu load average")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_LOAD_15m =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_15m_cpu_load")
                                    .help("zgc latest 15 minute cpu load average")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MMU_2MS =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_2ms_mmu_ratio")
                                    .help("zgc 2ms mmu ratio")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MMU_5MS =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_5ms_mmu_ratio")
                                    .help("zgc 5ms mmu ratio")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MMU_10MS =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_10ms_mmu_ratio")
                                    .help("zgc 10ms mmu ratio")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MMU_20MS =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_20ms_mmu_ratio")
                                    .help("zgc 20ms mmu ratio")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MMU_50MS =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_50ms_mmu_ratio")
                                    .help("zgc 50ms mmu ratio")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MMU_100MS =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_100ms_mmu_ratio")
                                    .help("zgc 100ms mmu ratio")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MARK_START_USED =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_mark_start_used_bytes")
                                    .help("zgc mark start used")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MARK_START_FREE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_mark_start_free_bytes")
                                    .help("zgc mark start free")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MARK_END_USED =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_mark_end_used_bytes")
                                    .help("zgc mark end used")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MARK_END_FREE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_mark_end_free_bytes")
                                    .help("zgc mark end free")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_RELOCATE_START_USED =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_relocate_start_used_bytes")
                                    .help("zgc relocate start used")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_RELOCATE_START_FREE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_relocate_start_free_bytes")
                                    .help("zgc relocate start free")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_RELOCATE_END_USED =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_relocate_end_used_bytes")
                                    .help("zgc relocate end used")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_RELOCATE_END_FREE =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_relocate_end_free_bytes")
                                    .help("zgc relocate end free")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_LIVE_MARK_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_live_mark_end_bytes")
                                    .help("zgc live mark end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_LIVE_RECLAIM_START =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_live_reclaim_start_bytes")
                                    .help("zgc live reclaim start")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_LIVE_RECLAIM_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_live_reclaim_end_bytes")
                                    .help("zgc live reclaim end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_ALLOCATED_MARK_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_allocated_mark_end_bytes")
                                    .help("zgc allocated mark end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_ALLOCATED_RECLAIM_START =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_allocated_reclaim_start_bytes")
                                    .help("zgc allocated reclaim start")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_ALLOCATED_RECLAIM_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_allocated_reclaim_end_bytes")
                                    .help("zgc allocated reclaim end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_GARBAGE_MARK_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_garbage_mark_end_bytes")
                                    .help("zgc garbage mark end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_GARBAGE_RECLAIM_START =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_garbage_reclaim_start_heap_bytes")
                                    .help("zgc garbage reclaim start")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_GARBAGE_RECLAIM_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_garbage_reclaim_end_bytes")
                                    .help("zgc garbage reclaim end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_RECLAIMED_RECLAIM_START =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_reclaimed_reclaim_start_bytes")
                                    .help("zgc reclaim start")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_RECLAIMED_RECLAIM_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_reclaimed_reclaim_end_bytes")
                                    .help("zgc reclaim end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MEMORY_RECLAIM_START =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_memory_reclaim_start_bytes")
                                    .help("zgc memory reclaim start")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_MEMORY_RECLAIM_END =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_memory_reclaim_end_bytes")
                                    .help("zgc memory reclaim end")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_METASPACE_USED =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_metaspace_used_bytes")
                                    .help("metaspace used memory")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_METASPACE_COMMITTED =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_metaspace_committed_bytes")
                                    .help("metaspace committed memory")
                                    .create());

    public static final CollectorProxy<Gauge.Child, Gauge> ZGC_METASPACE_RESERVED =
            CollectorProxy.of(
                    () ->
                            Gauge.build()
                                    .labelNames("path")
                                    .name("jgc_zgc_metaspace_reserved_bytes")
                                    .help("metaspace reserved memory")
                                    .create());

    private final List<CollectorProxy> collectors;

    private CollectorProxyRegistry() {
        super(true);
        this.collectors = new CopyOnWriteArrayList<>();
    }

    @Override
    public void register(Collector m) {
        super.register(m);
        if (m instanceof CollectorProxy) {
            collectors.add((CollectorProxy) m);
        }
    }

    @Override
    public void unregister(Collector m) {
        super.unregister(m);
        if (m instanceof CollectorProxy) {
            collectors.remove(m);
        }
    }

    public static void detach(Object object) {
        SINGLETON.collectors.forEach(collector -> collector.detach(object));
    }
}
