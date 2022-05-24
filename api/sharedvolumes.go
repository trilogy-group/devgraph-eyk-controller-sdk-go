package api

// Mount is the definition of PATCH /v2/apps/<app_id>/volumes/<volume-name>/sharedvolumes/<shared-volume-name>/path/.
type SharedMount struct {
	Values map[string]string `json:"values"`
}

// Unmount is the definition of PATCH /v2/apps/<app_id>/volumes/<volume-name>/sharedvolumes/<shared-volume-name>/path/.
type SharedUnmount struct {
	Values map[string]interface{} `json:"values"`
}

// Volume is the structure of an app's volume.
type SharedVolume struct {
	// Owner is the app owner.
	Owner string `json:"owner,omitempty"`
	// App is the app the tls settings apply to and cannot be updated.
	App string `json:"app,omitempty"`
	// ParentVolume is the volume where the mount point is copied from
	Parent_Volume string `json:"parentvolume,omitempty"`
	// ParentApp is the app the parent volume belongs to
	Parent_App string `json:"parentapp:omitempty"`
	// Created is the time that the volume was created and cannot be updated.
	Created string `json:"created,omitempty"`
	// Updated is the last time the TLS settings was changed and cannot be updated.
	Updated string `json:"updated,omitempty"`
	// UUID is a unique string reflecting the volume in its current state.
	// It changes every time the volume is changed and cannot be updated.
	UUID string `json:"uuid,omitempty"`
	// Volume's name
	Name string `json:"name,omitempty"`
	//Volume's size
	Size string `json:"size,omitempty"`
	// mount application's path
	Path map[string]interface{} `json:"path,omitempty"`
}

type SharedVolumes []SharedVolume
