package aggregation

import "regexp"

var pvcAggregationRegexps = map[int]*regexp.Regexp{}
var pvcAggregationLabelValues = map[int]string{}
var pvcAggregations = map[string]string{
	"storageclass.storage.k8s.io .* not found":                                                          "storageclass not found",
	"External provisioner is provisioning volume for claim .*":                                          "External provisioner is provisioning volume for claim",
	"Waiting for a volume to be created.*":                                                              "Waiting for a volume to be created",
	"Successfully provisioned volume .*":                                                                "Successfully provisioned volume",
	".*CSI migration enabled for .*; waiting for external resizer to expand the pvc.*":                  "CSI migration enabled for plugin; waiting for external resizer to expand the pvc",
	".*error getting CSI driver name for pvc .*, with error.*":                                          "Error getting CSI driver name for pvc",
	".*error setting resizer annotation to pvc .*, with error.*":                                        "Error setting resizer annotation to pvc",
	".*waiting for pod.* to be scheduled":                                                               "Waiting for pods to be scheduled",
	"Cannot bind to requested volume.*":                                                                 "Cannot bind to requested volume",
	".*volume.* already bound to a different claim.*":                                                   "Volume already bound to a different claim",
	"Cannot bind PersistentVolume.* to requested PersistentVolumeClaim due to incompatible volumeMode.": "Cannot bind PersistentVolume to requested PersistentVolumeClaim due to incompatible volumeMode.",
	".*plugin .* is not a CSI plugin. Only CSI plugin can provision a claim with a datasource":          "Plugin is not a CSI plugin",
	"Mount options are not supported by the provisioner but StorageClass .* has mount options .*":       "Mount options are not supported by the provisioner but StorageClass has mount options",
	"Failed to create provisioner.*":                                                                    "Failed to create provisioner",
	"Failed to get target node.*":                                                                       "Failed to get target node",
	"Failed to provision volume with StorageClass.*":                                                    "Failed to provision volume with StorageClass",
	"Error creating provisioned PV object for claim .* Deleting the volume.":                            "Error creating provisioned PV object for claim. Deleting the volume.",
	"Error cleaning provisioned volume for claim.* Please delete manually.":                             "Error cleaning provisioned volume for claim. Please delete manually.",
	"Successfully provisioned volume.* using.*":                                                         "Successfully provisioned volume using plugin",
	".*error getting CSI name for In tree plugin.*":                                                     "Error getting CSI name for In tree plugin",
	"Error saving claim.*":                                                                              "Error saving claim",
}

var pvAggregationRegexps = map[int]*regexp.Regexp{}
var pvAggregationLabelValues = map[int]string{}
var pvAggregations = map[string]string{
	"Cannot bind PersistentVolume to requested PersistentVolumeClaim .* due to incompatible volumeMode.": "Cannot bind PersistentVolume to requested PersistentVolumeClaim due to incompatible volumeMode.",
	"Recycle failed.*":                                                  "Recycle of volume failed",
	"Volume is used by pods.*":                                          "Volume is used by pods",
	".*failed to create deleter for volume.*":                           "Failed to create deleter for volume",
	".*persistent volume controller can't update finalizer.*":           "Persistent volume controller can't update finalizer",
	".*persistent Volume Controller can't anneal migration finalizer.*": "Persistent Volume Controller can't anneal migration finalizer",
	".*error getting deleter volume plugin for volume.*":                "Error getting deleter volume plugin for volume",
	"Recycler pod: .*":                                                  "Recycler pod",
	"rpc error: .* failed with error Bad request with.*":                "Rpc error: failed with error Bad request",
	"persistentvolume .* is still attached to node.*":                   "PersistentVolume is still attached to node",
}
