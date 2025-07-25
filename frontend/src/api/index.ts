import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'

type RoutingMethod = 'global' | 'direct'

export const DirectRoutingDomain = "https://direct.photos.parkers.wedding"
export const PubliclyRoutableDomain = "fallback-api.photos.parkers.wedding"

class ApiClient {
  private globalApi: AxiosInstance
  private directApi: AxiosInstance
  private directApiAvailable: boolean = false
  private lastDirectCheck: number = 0
  private directCheckInterval: number = 10000 // Check every 10 seconds

  constructor() {
    // Global API instance (fallback)
    this.globalApi = axios.create({
      baseURL: import.meta.env.VITE_API_BASE_URL || `https://${PubliclyRoutableDomain}`,
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json'
      },
      timeout: 10_000 // 10 seconds for global API
    })

    this.directApi = axios.create({
      baseURL: DirectRoutingDomain,
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json'
      },
      timeout: 3000 // Short timeout for direct connection check
    })
  }


  public async checkDirectApiAvailability(): Promise<boolean> {
    // Only check if enough time has passed
    const now = Date.now()
    if (now - this.lastDirectCheck < this.directCheckInterval) {
      return this.directApiAvailable
    }

    try {
      await this.directApi.get('/api/health', { timeout: 2000 })
      this.directApiAvailable = true
      this.lastDirectCheck = now
      return true
    } catch {
      this.directApiAvailable = false
      this.lastDirectCheck = now
      return false
    }
  }

  private async executeRequest<T>(
    method: 'get' | 'post' | 'put' | 'delete' | 'patch',
    url: string,
    data?: any,
    options?: { routingMethod?: RoutingMethod, config?: AxiosRequestConfig }
  ): Promise<AxiosResponse<T>> {
    // Check if direct API is available
    const canUseDirect = await this.checkDirectApiAvailability()

    if (canUseDirect && this.directApi && options?.routingMethod !== 'global') {
      try {
        // Try direct API first
        const response = await this.directApi[method]<T>(url, data, options?.config)
        return response
      } catch (error) {
        // If direct API fails, fall back to global API
        console.warn('Direct API failed, falling back to global API:', error)
        this.directApiAvailable = false
      }
    }

    if (options?.routingMethod === 'direct') {
      throw new Error('Direct API is not available')
    }

    if (method === 'get' || method === 'delete') {
      return await this.globalApi[method]<T>(url, options?.config)
    } else {
      return await this.globalApi[method]<T>(url, data, options?.config)
    }
  }

  // Public API methods
  async get<T = any>(url: string, options?: { routingMethod?: RoutingMethod, config?: AxiosRequestConfig }): Promise<AxiosResponse<T>> {
    return this.executeRequest<T>('get', url, undefined, options)
  }

  async post<T = any>(url: string, data?: any, options?: { routingMethod?: RoutingMethod, config?: AxiosRequestConfig }): Promise<AxiosResponse<T>> {
    return this.executeRequest<T>('post', url, data, options)
  }

  async put<T = any>(url: string, data?: any, options?: { routingMethod?: RoutingMethod, config?: AxiosRequestConfig }): Promise<AxiosResponse<T>> {
    return this.executeRequest<T>('put', url, data, options)
  }

  async delete<T = any>(url: string, options?: { routingMethod?: RoutingMethod, config?: AxiosRequestConfig }): Promise<AxiosResponse<T>> {
    return this.executeRequest<T>('delete', url, undefined, options)
  }

  async patch<T = any>(url: string, data?: any, options?: { routingMethod?: RoutingMethod, config?: AxiosRequestConfig }): Promise<AxiosResponse<T>> {
    return this.executeRequest<T>('patch', url, data, options)
  }

  // Get the current active endpoint (for static assets)
  async getActiveEndpoint(): Promise<string | null> {
    // Check availability to ensure we have the latest status
    const canUseDirect = await this.checkDirectApiAvailability()

    if (canUseDirect && this.directApi) {
      return this.directApi.defaults.baseURL || null
    }

    // Return null for relative URLs (no explicit base URL needed)
    return null
  }
}

// Create and export a singleton instance
const api = new ApiClient()
export default api