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

	HeapOccupancyBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_occupancy_before_collection_bytes",
		Help: "Heap occupancy (used) before collection",
	}, []string{"path", "host"})

	HeapOccupancyAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_occupancy_after_collection_bytes",
		Help: "Heap occupancy (used) after collection",
	}, []string{"path", "host"})

	HeapSizeBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_size_before_collection_bytes",
		Help: "Heap size before collection",
	}, []string{"path", "host"})

	HeapSizeAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_heap_size_after_collection_bytes",
		Help: "Heap size after collection",
	}, []string{"path", "host"})

	YoungOccupancyBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_young_occupancy_before_collection_bytes",
		Help: "Young generation occupancy before collection",
	}, []string{"path", "host"})

	YoungOccupancyAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_young_occupancy_after_collection_bytes",
		Help: "Young generation occupancy after collection",
	}, []string{"path", "host"})

	YoungSizeBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_young_size_before_collection_bytes",
		Help: "Young generation size before collection",
	}, []string{"path", "host"})

	YoungSizeAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_young_size_after_collection_bytes",
		Help: "Young generation size after collection",
	}, []string{"path", "host"})

	OldOccupancyBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_old_occupancy_before_collection_bytes",
		Help: "Old generation occupancy before collection",
	}, []string{"path", "host"})

	OldOccupancyAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_old_occupancy_after_collection_bytes",
		Help: "Old generation occupancy after collection",
	}, []string{"path", "host"})

	OldSizeBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_old_size_before_collection_bytes",
		Help: "Old generation size before collection",
	}, []string{"path", "host"})

	OldSizeAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_old_size_after_collection_bytes",
		Help: "Old generation size after collection",
	}, []string{"path", "host"})

	MetaspaceOccupancyBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_occupancy_before_collection_bytes",
		Help: "Metaspace occupancy before collection",
	}, []string{"path", "host"})

	MetaspaceOccupancyAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_occupancy_after_collection_bytes",
		Help: "Metaspace occupancy after collection",
	}, []string{"path", "host"})

	MetaspaceSizeBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_size_before_collection_bytes",
		Help: "Metaspace size before collection",
	}, []string{"path", "host"})

	MetaspaceSizeAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_metaspace_size_after_collection_bytes",
		Help: "Metaspace size after collection",
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
		Name: "jgc_zgc_1m_cpu_load",
		Help: "ZGC latest 1 minute CPU load average",
	}, []string{"path", "host"})

	ZGCLoad5m = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_5m_cpu_load",
		Help: "ZGC latest 5 minute CPU load average",
	}, []string{"path", "host"})

	ZGCLoad15m = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_15m_cpu_load",
		Help: "ZGC latest 15 minute CPU load average",
	}, []string{"path", "host"})

	ZGCMMU2ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_2ms_mmu_ratio",
		Help: "ZGC 2ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU5ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_5ms_mmu_ratio",
		Help: "ZGC 5ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU10ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_10ms_mmu_ratio",
		Help: "ZGC 10ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU20ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_20ms_mmu_ratio",
		Help: "ZGC 20ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU50ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_50ms_mmu_ratio",
		Help: "ZGC 50ms MMU ratio",
	}, []string{"path", "host"})

	ZGCMMU100ms = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_zgc_100ms_mmu_ratio",
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

	// GC Phase duration (unified: gc,phases lines)
	GCPhaseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_gc_phase_duration_seconds",
		Help: "Duration of individual GC phases (e.g., Marking Phase, Compaction Phase)",
	}, []string{"path", "host", "phase"})

	// GC Worker count
	GCWorkers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_gc_workers",
		Help: "Number of GC worker threads used",
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
		Name: "jgc_cms_symbol_and_string_table_process_seconds",
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

	// G1-specific (eden, survivor)
	G1EdenOccupancyBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_occupancy_before_collection_bytes",
		Help: "G1 Eden heap occupancy bytes before collection",
	}, []string{"path", "host"})

	G1EdenOccupancyAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_occupancy_after_collection_bytes",
		Help: "G1 Eden occupancy bytes after collection",
	}, []string{"path", "host"})

	G1EdenSizeBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_size_before_collection_bytes",
		Help: "G1 Eden size before collection",
	}, []string{"path", "host"})

	G1EdenSizeAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_size_after_collection_bytes",
		Help: "G1 Eden size after collection",
	}, []string{"path", "host"})

	G1SurvivorHeapOccupancyBeforeCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_heap_occupancy_before_collection_bytes",
		Help: "G1 survivor heap occupancy bytes before collection",
	}, []string{"path", "host"})

	G1SurvivorHeapOccupancyAfterCollection = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_heap_occupancy_after_collection_bytes",
		Help: "G1 survivor heap occupancy bytes after collection",
	}, []string{"path", "host"})

	G1SurvivorSizeBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_size_bytes",
		Help: "G1 survivor size",
	}, []string{"path", "host"})

	// G1 region counts
	G1EdenBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_before_collection_regions",
		Help: "Amount of G1 Eden region before collection",
	}, []string{"path", "host"})

	G1EdenAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_after_collection_regions",
		Help: "Amount of G1 Eden region after collection",
	}, []string{"path", "host"})

	G1EdenAssignRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_eden_assign_regions",
		Help: "Amount of G1 Eden assign regions",
	}, []string{"path", "host"})

	G1SurvivorBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_before_collection_regions",
		Help: "Amount of G1 survivor region before collection",
	}, []string{"path", "host"})

	G1SurvivorAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_after_collection_regions",
		Help: "Amount of G1 survivor region after collection",
	}, []string{"path", "host"})

	G1SurvivorAssignRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_survivor_assign_regions",
		Help: "Amount of G1 survivor assign regions",
	}, []string{"path", "host"})

	G1OldBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_old_before_collection_regions",
		Help: "Amount of G1 old regions before collection",
	}, []string{"path", "host"})

	G1OldAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_old_after_collection_regions",
		Help: "Amount of G1 old regions after collection",
	}, []string{"path", "host"})

	G1HumongousBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_humongous_before_collection_regions",
		Help: "Amount of G1 humongous regions before collection",
	}, []string{"path", "host"})

	G1HumongousAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_humongous_after_collection_regions",
		Help: "Amount of G1 humongous regions after collection",
	}, []string{"path", "host"})

	G1ArchiveBeforeCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_archive_before_collection_regions",
		Help: "Amount of G1 archive regions before collection",
	}, []string{"path", "host"})

	G1ArchiveAfterCollectionRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_archive_after_collection_regions",
		Help: "Amount of G1 archive regions after collection",
	}, []string{"path", "host"})

	G1ArchiveAssignRegions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_g1_archive_assign_regions",
		Help: "Amount of G1 archive assign regions",
	}, []string{"path", "host"})

	// Reference processing
	SoftReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_soft_references",
		Help: "Amount of soft references",
	}, []string{"path", "host"})

	SoftReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_soft_reference_pause_duration_seconds",
		Help: "Soft reference pause duration",
	}, []string{"path", "host"})

	WeakReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_weak_references",
		Help: "Amount of weak references",
	}, []string{"path", "host"})

	WeakReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_weak_reference_pause_seconds",
		Help: "Weak reference pause duration",
	}, []string{"path", "host"})

	FinalReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_final_references",
		Help: "Amount of final references",
	}, []string{"path", "host"})

	FinalReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_final_reference_pause_duration_seconds",
		Help: "Final reference pause duration",
	}, []string{"path", "host"})

	PhantomReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_phantom_references",
		Help: "Amount of phantom references",
	}, []string{"path", "host"})

	FreePhantomReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_free_phantom_references",
		Help: "Amount of free phantom references",
	}, []string{"path", "host"})

	PhantomReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_phantom_reference_pause_duration_seconds",
		Help: "Phantom reference pause duration",
	}, []string{"path", "host"})

	JNIWeakReferences = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_jni_weak_references",
		Help: "Amount of JNI weak references",
	}, []string{"path", "host"})

	JNIWeakReferencePauseDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "jgc_jni_weak_reference_pause_duration_seconds",
		Help: "JNI weak reference pause duration",
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
		HeapOccupancyBeforeCollection, HeapOccupancyAfterCollection,
		HeapSizeBeforeCollection, HeapSizeAfterCollection,
		YoungOccupancyBeforeCollection, YoungOccupancyAfterCollection,
		YoungSizeBeforeCollection, YoungSizeAfterCollection,
		OldOccupancyBeforeCollection, OldOccupancyAfterCollection,
		OldSizeBeforeCollection, OldSizeAfterCollection,
		MetaspaceOccupancyBeforeCollection, MetaspaceOccupancyAfterCollection,
		MetaspaceSizeBeforeCollection, MetaspaceSizeAfterCollection,
		ZGCPauseMarkStartDuration, ZGCConcurrentMarkDuration,
		ZGCPauseMarkEndDuration, ZGCPauseRelocateStartDuration,
		ZGCConcurrentRelocateDuration,
		ZGCLoad1m, ZGCLoad5m, ZGCLoad15m,
		ZGCMMU2ms, ZGCMMU5ms, ZGCMMU10ms, ZGCMMU20ms, ZGCMMU50ms, ZGCMMU100ms,
		ZGCMetaspaceUsed, ZGCMetaspaceCommitted,
		SafepointDuration, SafepointStopThreadsDuration,
		GCPhaseDuration, GCWorkers,
		CMSSymbolTableProcessDuration, CMSStringTableProcessDuration,
		CMSClassUnloadingProcessDuration, CMSSymbolAndStringTableProcessDuration,
		ZGCConcurrentMarkFreeDuration, ZGCProcessNonStrongReferencesDuration,
		ZGCConcurrentResetRelocationsetDuration, ZGCConcurrentSelectRelocationsetDuration,
		G1EdenOccupancyBeforeCollection, G1EdenOccupancyAfterCollection,
		G1EdenSizeBeforeCollection, G1EdenSizeAfterCollection,
		G1SurvivorHeapOccupancyBeforeCollection, G1SurvivorHeapOccupancyAfterCollection,
		G1SurvivorSizeBytes,
		G1EdenBeforeCollectionRegions, G1EdenAfterCollectionRegions, G1EdenAssignRegions,
		G1SurvivorBeforeCollectionRegions, G1SurvivorAfterCollectionRegions, G1SurvivorAssignRegions,
		G1OldBeforeCollectionRegions, G1OldAfterCollectionRegions,
		G1HumongousBeforeCollectionRegions, G1HumongousAfterCollectionRegions,
		G1ArchiveBeforeCollectionRegions, G1ArchiveAfterCollectionRegions, G1ArchiveAssignRegions,
		SoftReferences, SoftReferencePauseDuration,
		WeakReferences, WeakReferencePauseDuration,
		FinalReferences, FinalReferencePauseDuration,
		PhantomReferences, FreePhantomReferences, PhantomReferencePauseDuration,
		JNIWeakReferences, JNIWeakReferencePauseDuration,
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
