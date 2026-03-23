package parser

import "time"

type GCType string

const (
	GCTypeG1       GCType = "G1"
	GCTypeCMS      GCType = "CMS"
	GCTypeZGC      GCType = "ZGC"
	GCTypeParallel GCType = "Parallel"
	GCTypeSerial   GCType = "Serial"
	GCTypeUnknown  GCType = "Unknown"
)

type GCEvent struct {
	Timestamp time.Time
	GCType    GCType
	Category  string
	Duration  float64 // seconds
	IsPause   bool

	HeapBeforeKB  int64
	HeapAfterKB   int64
	HeapTotalKB   int64
	YoungBeforeKB int64
	YoungAfterKB  int64
	YoungTotalKB  int64
	OldBeforeKB   int64
	OldAfterKB    int64
	OldTotalKB    int64
	MetaBeforeKB  int64
	MetaAfterKB   int64
	MetaTotalKB   int64

	// G1-specific
	G1EdenBeforeKB     int64
	G1EdenAfterKB      int64
	G1EdenTotalKB      int64
	G1SurvivorBeforeKB int64
	G1SurvivorAfterKB  int64

	// G1 region counts (unified log format)
	G1EdenRegionBefore      int64
	G1EdenRegionAfter       int64
	G1EdenRegionAssign      int64
	G1SurvivorRegionBefore  int64
	G1SurvivorRegionAfter   int64
	G1SurvivorRegionAssign  int64
	G1OldRegionBefore       int64
	G1OldRegionAfter        int64
	G1HumongousRegionBefore int64
	G1HumongousRegionAfter  int64
	G1ArchiveRegionBefore   int64
	G1ArchiveRegionAfter    int64
	G1ArchiveRegionAssign   int64

	// ZGC-specific: phase durations (ms)
	ZGCPauseMarkStartMs         float64
	ZGCConcurrentMarkMs         float64
	ZGCPauseMarkEndMs           float64
	ZGCPauseRelocateMs          float64
	ZGCConcurrentRelocMs        float64
	ZGCConcurrentMarkFreeMs     float64
	ZGCConcurrentProcessNSRMs   float64 // Non-Strong References
	ZGCConcurrentResetRelocMs   float64
	ZGCConcurrentSelectRelocMs  float64
	ZGCConcurrentRemapRootsMs   float64 // Generational ZGC old-gen only
	ZGCConcurrentMarkContinueMs float64 // Generational ZGC

	ZGCCollectionType string // "minor" / "major" / "" (non-generational)

	ZGCUsedMB              int64
	ZGCFreeMB              int64
	ZGCLoad1m              float64
	ZGCLoad5m              float64
	ZGCLoad15m             float64
	ZGCMMU2ms              float64
	ZGCMMU5ms              float64
	ZGCMMU10ms             float64
	ZGCMMU20ms             float64
	ZGCMMU50ms             float64
	ZGCMMU100ms            float64
	ZGCMetaspaceUsedKB     int64
	ZGCMetaspaceCommitKB   int64
	ZGCMetaspaceReservedKB int64

	// ZGC heap memory details (from gc,heap Used/Free lines)
	ZGCMarkStartUsedKB     int64
	ZGCMarkStartFreeKB     int64
	ZGCMarkEndUsedKB       int64
	ZGCMarkEndFreeKB       int64
	ZGCRelocateStartUsedKB int64
	ZGCRelocateStartFreeKB int64
	ZGCRelocateEndUsedKB   int64
	ZGCRelocateEndFreeKB   int64

	// CMS-specific
	CMSClassUnloadingMs float64
	CMSSymbolTableMs    float64
	CMSStringTableMs    float64

	// Reference processing (Soft/Weak/Final/Phantom/JNI)
	SoftRefCount      int64
	SoftRefPauseMs    float64
	WeakRefCount      int64
	WeakRefPauseMs    float64
	FinalRefCount     int64
	FinalRefPauseMs   float64
	PhantomRefCount   int64
	PhantomRefFree    int64
	PhantomRefPauseMs float64
	JNIWeakRefCount   int64
	JNIWeakRefPauseMs float64

	// GC cause (if available)
	Cause string
}

type EventHandler interface {
	Handle(event *GCEvent)
}

type Parser interface {
	Feed(line string)
	Flush()
	SetHandler(handler EventHandler)
	GCType() GCType
}
