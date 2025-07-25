export interface User {
  id: string
  name: string
  createdAt: string
}

export interface Photo {
  id: string
  originalName: string
  mimeType: string
  uploadedAt: string
  uploadedBy: string
  uploaderName?: string
}

export interface NetworkRoutingHint {
  wifi: {
    ssid: string
    password: string
  }
}
