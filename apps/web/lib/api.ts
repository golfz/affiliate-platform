const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

export interface CreateProductRequest {
  source: string;
  sourceType: 'url' | 'sku';
  lazada_url?: string;
  shopee_url?: string;
}

export interface ProductResponse {
  id: string;
  title: string;
  image_url: string;
  offers?: OfferResponse[];
  created_at: string;
}

export interface OfferResponse {
  id: string;
  marketplace: string;
  store_name: string;
  price: number;
  last_checked_at: string;
}

export interface ProductOffersResponse {
  product_id: string;
  offers: OfferResponse[];
  best_price?: {
    marketplace: string;
    price: number;
  };
}

export interface CreateCampaignRequest {
  name: string;
  utm_campaign: string;
  start_at: string;
  end_at: string;
  product_ids?: string[];
}

export interface UpdateCampaignRequest {
  name?: string;
  utm_campaign?: string;
  start_at?: string;
  end_at?: string;
  product_ids?: string[];
}

export interface CampaignResponse {
  id: string;
  name: string;
  utm_campaign: string;
  start_at: string;
  end_at: string;
  created_at: string;
  product_ids?: string[]; // Product IDs in this campaign
}

export interface CampaignPublicResponse {
  id: string;
  name: string;
  start_at: string;
  end_at: string;
  products: CampaignProduct[];
}

export interface CampaignProduct {
  id: string;
  title: string;
  image_url: string;
  offers: OfferResponse[];
  best_price?: {
    marketplace: string;
    price: number;
  };
  links?: ProductLink[];
}

export interface ProductLink {
  marketplace: string;
  short_code: string;
  full_url: string;
}

export interface CreateLinkRequest {
  product_id: string;
  campaign_id: string;
  marketplace: 'lazada' | 'shopee';
}

export interface LinkResponse {
  id: string;
  product_id: string;
  campaign_id: string;
  marketplace: string;
  short_code: string;
  target_url: string;
  full_url: string;
  created_at: string;
}

export interface ErrorResponse {
  error: string;
  message: string;
  code: string;
}

async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  const response = await fetch(url, {
    ...options,
    headers,
    // credentials: 'include', // Not needed since AllowCredentials: false
    mode: 'cors', // Enable CORS
  });

  if (!response.ok) {
    const error: ErrorResponse = await response.json().catch(() => ({
      error: 'Unknown Error',
      message: `HTTP ${response.status}: ${response.statusText}`,
      code: 'HTTP_ERROR',
    }));
    throw new Error(error.message || error.error);
  }

  // Handle 204 No Content (no response body)
  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

// Product API
export async function createProduct(data: CreateProductRequest): Promise<ProductResponse> {
  return apiRequest<ProductResponse>('/api/products', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getAllProducts(limit?: number, offset?: number): Promise<ProductResponse[]> {
  const params = new URLSearchParams();
  if (limit) params.append('limit', limit.toString());
  if (offset) params.append('offset', offset.toString());
  const query = params.toString();
  return apiRequest<ProductResponse[]>(`/api/products${query ? `?${query}` : ''}`);
}

export async function getProductOffers(productId: string): Promise<ProductOffersResponse> {
  return apiRequest<ProductOffersResponse>(`/api/products/${productId}/offers`);
}

export async function deleteProduct(productId: string): Promise<void> {
  return apiRequest<void>(`/api/products/${productId}`, {
    method: 'DELETE',
  });
}

// Campaign API
export async function getAllCampaigns(limit?: number, offset?: number): Promise<CampaignResponse[]> {
  const params = new URLSearchParams();
  if (limit) params.append('limit', limit.toString());
  if (offset) params.append('offset', offset.toString());
  const query = params.toString();
  return apiRequest<CampaignResponse[]>(`/api/campaigns${query ? `?${query}` : ''}`);
}

export async function getCampaign(campaignId: string): Promise<CampaignResponse> {
  return apiRequest<CampaignResponse>(`/api/campaigns/${campaignId}`);
}

export async function createCampaign(data: CreateCampaignRequest): Promise<CampaignResponse> {
  return apiRequest<CampaignResponse>('/api/campaigns', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getPublicCampaign(campaignId: string): Promise<CampaignPublicResponse> {
  return apiRequest<CampaignPublicResponse>(`/api/campaigns/${campaignId}/public`);
}

export async function updateCampaign(campaignId: string, data: UpdateCampaignRequest): Promise<CampaignResponse> {
  return apiRequest<CampaignResponse>(`/api/campaigns/${campaignId}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  });
}

export async function updateCampaignProducts(campaignId: string, productIds: string[]): Promise<CampaignResponse> {
  return apiRequest<CampaignResponse>(`/api/campaigns/${campaignId}/products`, {
    method: 'PATCH',
    body: JSON.stringify({ product_ids: productIds }),
  });
}

export async function deleteCampaign(campaignId: string): Promise<void> {
  return apiRequest<void>(`/api/campaigns/${campaignId}`, {
    method: 'DELETE',
  });
}

// Link API
export async function createLink(data: CreateLinkRequest): Promise<LinkResponse> {
  return apiRequest<LinkResponse>('/api/links', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// Helper to get redirect URL
export function getRedirectUrl(shortCode: string): string {
  return `${API_BASE_URL}/go/${shortCode}`;
}
