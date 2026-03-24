package metric

import "github.com/prometheus/client_golang/prometheus"

var (
	EventDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_event_duration_seconds",
		Help: "Duration of GC events",
	}, []string{"path", "host", "category"})

	EventPauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_event_pause_duration_seconds",
		Help: "Duration of GC pause events",
	}, []string{"path", "host", "category"})

	EventLastMinDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "jgc_event_last_minute_duration_seconds",
		Help:       "Last minute GC event duration",
		MaxAge:     60e9, // 60 seconds
		AgeBuckets: 6,
		Objectives: map[float64]float64{0: 0.05, 0.5: 0.05, 0.75: 0.05, 1.0: 0.05},
	}, []string{"path", "host"})

	EventLastMinPauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "jgc_event_last_minute_pause_duration_seconds",
		Help:       "Last minute GC pause duration",
		MaxAge:     60e9,
		AgeBuckets: 6,
		Objectives: map[float64]float64{0: 0.05, 0.5: 0.05, 0.75: 0.05, 1.0: 0.05},
	}, []string{"path", "host"})

	HeapUsedBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_used_before_collection_bytes",
		Help: "Heap used bytes before collection (from GC log used/total)",
	}, []string{"path", "host"})

	HeapUsedAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_used_after_collection_bytes",
		Help: "Heap used bytes after collection (from GC log used/total)",
	}, []string{"path", "host"})

	HeapSizeBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_size_before_collection_bytes",
		Help: "Heap committed capacity in bytes at collection (total from GC log; distinct from used bytes)",
	}, []string{"path", "host"})

	HeapSizeAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_size_after_collection_bytes",
		Help: "Heap committed capacity in bytes after collection (same total as before when log reports single heap size)",
	}, []string{"path", "host"})

	MetaspaceUsedBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_used_before_collection_bytes",
		Help: "Metaspace used bytes before collection",
	}, []string{"path", "host"})

	MetaspaceUsedAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_used_after_collection_bytes",
		Help: "Metaspace used bytes after collection",
	}, []string{"path", "host"})

	MetaspaceSizeBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_size_before_collection_bytes",
		Help: "Metaspace committed capacity before collection (distinct from used bytes)",
	}, []string{"path", "host"})

	MetaspaceSizeAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_size_after_collection_bytes",
		Help: "Metaspace committed capacity after collection (distinct from used bytes)",
	}, []string{"path", "host"})

	// ZGC-specific
	ZGCPauseMarkStartDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_pause_mark_start_duration_seconds",
		Help: "ZGC pause mark start duration",
	}, []string{"path", "host"})

	ZGCConcurrentMarkDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_concurrent_mark_duration_seconds",
		Help: "ZGC concurrent mark duration",
	}, []string{"path", "host"})

	ZGCPauseMarkEndDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_pause_mark_end_duration_seconds",
		Help: "ZGC pause mark end duration",
	}, []string{"path", "host"})

	ZGCPauseRelocateStartDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_pause_relocate_start_duration_seconds",
		Help: "ZGC pause relocate start duration",
	}, []string{"path", "host"})

	ZGCConcurrentRelocateDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_concurrent_relocate_duration_seconds",
		Help: "ZGC concurrent relocate duration",
	}, []string{"path", "host"})

	ZGCLoad1m = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_cpu_load_1m",
		Help: "ZGC latest 1 minute CPU load average",
	}, []string{"path", "host"})

	ZGCLoad5m = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_cpu_load_5m",
		Help: "ZGC latest 5 minute CPU load average",
	}, []string{"path", "host"})

	ZGCLoad15m = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_cpu_load_15m",
		Help: "ZGC latest 15 minute CPU load average",
	}, []string{"path", "host"})

	ZGCMMU2ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mmu_ratio_2ms",
		Help: "ZGC 2ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU5ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mmu_ratio_5ms",
		Help: "ZGC 5ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU10ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mmu_ratio_10ms",
		Help: "ZGC 10ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU20ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mmu_ratio_20ms",
		Help: "ZGC 20ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU50ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mmu_ratio_50ms",
		Help: "ZGC 50ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU100ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mmu_ratio_100ms",
		Help: "ZGC 100ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMetaspaceUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_metaspace_used_bytes",
		Help: "ZGC metaspace used",
	}, []string{"path", "host"})

	ZGCMetaspaceCommitted = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_metaspace_committed_bytes",
		Help: "ZGC metaspace committed",
	}, []string{"path", "host"})

	// Safepoint metrics
	SafepointDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_safepoint_duration_seconds",
		Help: "Total time application threads were stopped at safepoint",
	}, []string{"path", "host"})

	SafepointStopThreadsDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_safepoint_stop_threads_seconds",
		Help: "Time taken to stop application threads for safepoint",
	}, []string{"path", "host"})

	// GC Worker count (unified JVM log [gc,task] lines; not collector-specific)
	GCWorkers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_gc_workers",
		Help: "GC worker thread count)",
	}, []string{"path", "host"})

	// CMS String/Symbol table durations (from CMS Remark or unified gc,phases)
	CMSSymbolTableProcessDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_symbol_table_process_duration_seconds",
		Help: "Symbol table process time",
	}, []string{"path", "host"})

	CMSStringTableProcessDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_string_table_process_duration_seconds",
		Help: "String table process duration",
	}, []string{"path", "host"})

	CMSClassUnloadingProcessDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_class_unloading_process_duration_seconds",
		Help: "Class unloading process time",
	}, []string{"path", "host"})

	CMSSymbolAndStringTableProcessDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_symbol_and_string_table_process_duration_seconds",
		Help: "Symbol and string table process duration",
	}, []string{"path", "host"})

	// ZGC additional phase durations
	ZGCConcurrentMarkFreeDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_concurrent_mark_free_duration_seconds",
		Help: "ZGC concurrent mark free duration",
	}, []string{"path", "host"})

	ZGCProcessNonStrongReferencesDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_process_non_strong_references_duration_seconds",
		Help: "ZGC process non-strong references duration",
	}, []string{"path", "host"})

	ZGCConcurrentResetRelocationsetDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_concurrent_reset_relocationset_duration_seconds",
		Help: "ZGC concurrent reset relocationset duration",
	}, []string{"path", "host"})

	ZGCConcurrentSelectRelocationsetDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_zgc_concurrent_select_relocationset_duration_seconds",
		Help: "ZGC concurrent select relocationset duration",
	}, []string{"path", "host"})

	// G1-specific: region counts (unified log); no generic young/old gauges.
	G1EdenBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_heap_before_collection_regions",
		Help: "G1 Eden heap regions before collection",
	}, []string{"path", "host"})

	G1EdenAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_heap_after_collection_regions",
		Help: "G1 Eden heap regions after collection",
	}, []string{"path", "host"})

	G1EdenAssignRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_heap_assign_regions",
		Help: "G1 Eden heap assign region max",
	}, []string{"path", "host"})

	G1SurvivorBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_heap_before_collection_regions",
		Help: "G1 survivor heap regions before collection",
	}, []string{"path", "host"})

	G1SurvivorAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_heap_after_collection_regions",
		Help: "G1 survivor heap regions after collection",
	}, []string{"path", "host"})

	G1SurvivorAssignRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_heap_assign_regions",
		Help: "G1 survivor heap assign region max",
	}, []string{"path", "host"})

	G1OldBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_old_heap_before_collection_regions",
		Help: "G1 old heap regions before collection",
	}, []string{"path", "host"})

	G1OldAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_old_heap_after_collection_regions",
		Help: "G1 old heap regions after collection",
	}, []string{"path", "host"})

	G1HumongousBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_humongous_heap_before_collection_regions",
		Help: "G1 humongous heap regions before collection",
	}, []string{"path", "host"})

	G1HumongousAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_humongous_heap_after_collection_regions",
		Help: "G1 humongous heap regions after collection",
	}, []string{"path", "host"})

	G1ArchiveBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_archive_heap_before_collection_regions",
		Help: "G1 archive heap regions before collection",
	}, []string{"path", "host"})

	G1ArchiveAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_archive_heap_after_collection_regions",
		Help: "G1 archive heap regions after collection",
	}, []string{"path", "host"})

	G1ArchiveAssignRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_archive_heap_assign_regions",
		Help: "G1 archive heap assign region max",
	}, []string{"path", "host"})

	// Reference processing: G1 Remark / CMS Remark with PrintReferenceGC (parsers in g1.go, cms.go only)
	G1SoftReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_soft_references",
		Help: "G1 soft reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	G1SoftReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_g1_soft_reference_pause_duration_seconds",
		Help: "G1 soft reference processing pause (remark)",
	}, []string{"path", "host"})

	CMSSoftReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_cms_soft_references",
		Help: "CMS soft reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	CMSSoftReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_soft_reference_pause_duration_seconds",
		Help: "CMS soft reference processing pause (remark)",
	}, []string{"path", "host"})

	G1WeakReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_weak_references",
		Help: "G1 weak reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	G1WeakReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_g1_weak_reference_pause_duration_seconds",
		Help: "G1 weak reference processing pause (remark)",
	}, []string{"path", "host"})

	CMSWeakReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_cms_weak_references",
		Help: "CMS weak reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	CMSWeakReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_weak_reference_pause_duration_seconds",
		Help: "CMS weak reference processing pause (remark)",
	}, []string{"path", "host"})

	G1FinalReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_final_references",
		Help: "G1 final reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	G1FinalReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_g1_final_reference_pause_duration_seconds",
		Help: "G1 final reference processing pause (remark)",
	}, []string{"path", "host"})

	CMSFinalReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_cms_final_references",
		Help: "CMS final reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	CMSFinalReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_final_reference_pause_duration_seconds",
		Help: "CMS final reference processing pause (remark)",
	}, []string{"path", "host"})

	G1PhantomReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_phantom_references",
		Help: "G1 phantom reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	G1FreePhantomReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_free_phantom_references",
		Help: "G1 free phantom references from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	G1PhantomReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_g1_phantom_reference_pause_duration_seconds",
		Help: "G1 phantom reference processing pause (remark)",
	}, []string{"path", "host"})

	CMSPhantomReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_cms_phantom_references",
		Help: "CMS phantom reference count from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	CMSFreePhantomReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_cms_free_phantom_references",
		Help: "CMS free phantom references from remark output (PrintReferenceGC)",
	}, []string{"path", "host"})

	CMSPhantomReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_phantom_reference_pause_duration_seconds",
		Help: "CMS phantom reference processing pause (remark)",
	}, []string{"path", "host"})

	G1JNIWeakReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_jni_weak_references",
		Help: "G1 JNI weak reference slot (count may be unset depending on log line)",
	}, []string{"path", "host"})

	G1JNIWeakReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_g1_jni_weak_reference_pause_duration_seconds",
		Help: "G1 JNI weak reference processing pause (remark)",
	}, []string{"path", "host"})

	CMSJNIWeakReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_cms_jni_weak_references",
		Help: "CMS JNI weak reference slot (count may be unset depending on log line)",
	}, []string{"path", "host"})

	CMSJNIWeakReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_cms_jni_weak_reference_pause_duration_seconds",
		Help: "CMS JNI weak reference processing pause (remark)",
	}, []string{"path", "host"})

	// ZGC heap memory details
	ZGCMarkStartUsedBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mark_start_used_bytes",
		Help: "ZGC mark start used",
	}, []string{"path", "host"})

	ZGCMarkStartFreeBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mark_start_free_bytes",
		Help: "ZGC mark start free",
	}, []string{"path", "host"})

	ZGCMarkEndUsedBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mark_end_used_bytes",
		Help: "ZGC mark end used",
	}, []string{"path", "host"})

	ZGCMarkEndFreeBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_mark_end_free_bytes",
		Help: "ZGC mark end free",
	}, []string{"path", "host"})

	ZGCRelocateStartUsedBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_relocate_start_used_bytes",
		Help: "ZGC relocate start used",
	}, []string{"path", "host"})

	ZGCRelocateStartFreeBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_relocate_start_free_bytes",
		Help: "ZGC relocate start free",
	}, []string{"path", "host"})

	ZGCRelocateEndUsedBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_relocate_end_used_bytes",
		Help: "ZGC relocate end used",
	}, []string{"path", "host"})

	ZGCRelocateEndFreeBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_relocate_end_free_bytes",
		Help: "ZGC relocate end free",
	}, []string{"path", "host"})

	ZGCMetaspaceReservedBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_metaspace_reserved_bytes",
		Help: "ZGC metaspace reserved memory",
	}, []string{"path", "host"})

	// Throughput and allocation rate (updated periodically from dashboard)
	ThroughputGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_throughput_ratio",
		Help: "GC throughput ratio (0-1), 1 - pause_time/total_time",
	}, []string{"path", "host"})

	AllocationRateGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_allocation_rate_mb_per_sec",
		Help: "Heap allocation rate between consecutive GCs (MB/s)",
	}, []string{"path", "host"})
)

// AllVecMetrics returns all GC Vec metrics for bulk operations (e.g. cleanup).
func AllVecMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		EventDuration, EventPauseDuration,
		EventLastMinDuration, EventLastMinPauseDuration,
		HeapUsedBeforeCollection, HeapUsedAfterCollection,
		HeapSizeBeforeCollection, HeapSizeAfterCollection,
		MetaspaceUsedBeforeCollection, MetaspaceUsedAfterCollection,
		MetaspaceSizeBeforeCollection, MetaspaceSizeAfterCollection,
		ZGCPauseMarkStartDuration, ZGCConcurrentMarkDuration,
		ZGCPauseMarkEndDuration, ZGCPauseRelocateStartDuration,
		ZGCConcurrentRelocateDuration,
		ZGCLoad1m, ZGCLoad5m, ZGCLoad15m,
		ZGCMMU2ms, ZGCMMU5ms, ZGCMMU10ms, ZGCMMU20ms, ZGCMMU50ms, ZGCMMU100ms,
		ZGCMetaspaceUsed, ZGCMetaspaceCommitted,
		SafepointDuration, SafepointStopThreadsDuration,
		GCWorkers,
		CMSSymbolTableProcessDuration, CMSStringTableProcessDuration,
		CMSClassUnloadingProcessDuration, CMSSymbolAndStringTableProcessDuration,
		ZGCConcurrentMarkFreeDuration, ZGCProcessNonStrongReferencesDuration,
		ZGCConcurrentResetRelocationsetDuration, ZGCConcurrentSelectRelocationsetDuration,
		G1EdenBeforeCollectionRegions, G1EdenAfterCollectionRegions, G1EdenAssignRegions,
		G1SurvivorBeforeCollectionRegions, G1SurvivorAfterCollectionRegions, G1SurvivorAssignRegions,
		G1OldBeforeCollectionRegions, G1OldAfterCollectionRegions,
		G1HumongousBeforeCollectionRegions, G1HumongousAfterCollectionRegions,
		G1ArchiveBeforeCollectionRegions, G1ArchiveAfterCollectionRegions, G1ArchiveAssignRegions,
		G1SoftReferences, G1SoftReferencePauseDuration,
		CMSSoftReferences, CMSSoftReferencePauseDuration,
		G1WeakReferences, G1WeakReferencePauseDuration,
		CMSWeakReferences, CMSWeakReferencePauseDuration,
		G1FinalReferences, G1FinalReferencePauseDuration,
		CMSFinalReferences, CMSFinalReferencePauseDuration,
		G1PhantomReferences, G1FreePhantomReferences, G1PhantomReferencePauseDuration,
		CMSPhantomReferences, CMSFreePhantomReferences, CMSPhantomReferencePauseDuration,
		G1JNIWeakReferences, G1JNIWeakReferencePauseDuration,
		CMSJNIWeakReferences, CMSJNIWeakReferencePauseDuration,
		ZGCMarkStartUsedBytes, ZGCMarkStartFreeBytes,
		ZGCMarkEndUsedBytes, ZGCMarkEndFreeBytes,
		ZGCRelocateStartUsedBytes, ZGCRelocateStartFreeBytes,
		ZGCRelocateEndUsedBytes, ZGCRelocateEndFreeBytes,
		ZGCMetaspaceReservedBytes,
		ThroughputGauge, AllocationRateGauge,
	}
}

func init() {
	for _, m := range AllVecMetrics() {
		Registry.MustRegister(m)
	}
}
