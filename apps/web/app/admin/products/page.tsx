'use client'

import { useState, useEffect } from 'react'
import AdminLayout from '@/components/AdminLayout'
import { createProduct, getAllProducts, deleteProduct, type ProductResponse } from '@/lib/api'

export default function ProductsPage() {
  const [lazadaUrl, setLazadaUrl] = useState('')
  const [shopeeUrl, setShopeeUrl] = useState('')
  const [productTitle, setProductTitle] = useState('') // For future use - currently disabled
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [urlErrors, setUrlErrors] = useState<{ lazada?: string; shopee?: string }>({})
  const [products, setProducts] = useState<ProductResponse[]>([])
  const [loadingProducts, setLoadingProducts] = useState(true)
  const [deletingProductId, setDeletingProductId] = useState<string | null>(null)

  // Validate URL matches platform
  const validateUrl = (url: string, platform: 'lazada' | 'shopee'): string | null => {
    if (!url) return null // Empty URL is OK (optional)
    
    try {
      const urlObj = new URL(url.toLowerCase())
      const hostname = urlObj.hostname.toLowerCase()
      
      if (platform === 'lazada') {
        if (!hostname.includes('lazada')) {
          return 'URL must be from Lazada (lazada.co.th, lazada.com, etc.)'
        }
      } else if (platform === 'shopee') {
        if (!hostname.includes('shopee')) {
          return 'URL must be from Shopee (shopee.co.th, shopee.com, etc.)'
        }
      }
      
      return null
    } catch {
      return 'Invalid URL format'
    }
  }

  const handleLazadaUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setLazadaUrl(value)
    const error = validateUrl(value, 'lazada')
    setUrlErrors(prev => ({ ...prev, lazada: error || undefined }))
  }

  const handleShopeeUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setShopeeUrl(value)
    const error = validateUrl(value, 'shopee')
    setUrlErrors(prev => ({ ...prev, shopee: error || undefined }))
  }

  // Fetch products on mount
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setLoadingProducts(true)
        const productsList = await getAllProducts()
        setProducts(productsList)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch products')
      } finally {
        setLoadingProducts(false)
      }
    }

    fetchProducts()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    // Validate at least one URL is provided
    if (!lazadaUrl && !shopeeUrl) {
      setError('Please provide at least one URL (Lazada or Shopee)')
      setLoading(false)
      return
    }

    // Validate URLs
    const lazadaError = validateUrl(lazadaUrl, 'lazada')
    const shopeeError = validateUrl(shopeeUrl, 'shopee')
    
    if (lazadaError || shopeeError) {
      setUrlErrors({
        lazada: lazadaError || undefined,
        shopee: shopeeError || undefined,
      })
      setError('Please fix URL validation errors')
      setLoading(false)
      return
    }

    try {
      // Determine primary URL (prefer Lazada if both are provided)
      const primaryUrl = lazadaUrl || shopeeUrl
      
      // Create product with both URLs if provided
      const product = await createProduct({
        source: primaryUrl,
        sourceType: 'url',
        lazada_url: lazadaUrl || undefined,
        shopee_url: shopeeUrl || undefined,
      })
      
      // Refresh products list to get offers
      const productsList = await getAllProducts()
      setProducts(productsList)
      // Clear form
      setLazadaUrl('')
      setShopeeUrl('')
      setUrlErrors({})
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create product')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteProduct = async (productId: string) => {
    if (!confirm('Are you sure you want to delete this product? This will also delete all related offers, links, and clicks.')) {
      return
    }

    try {
      setDeletingProductId(productId)
      setError(null)
      await deleteProduct(productId)
      // Remove from list
      setProducts(products.filter(p => p.id !== productId))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete product')
    } finally {
      setDeletingProductId(null)
    }
  }

  return (
    <AdminLayout>
      <div className="px-4 sm:px-0">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Products</h1>

        {/* Add Product Form */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900">Add Product</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="product_title" className="block text-sm font-medium text-gray-700 mb-2">
                Product Title <span className="text-xs text-gray-500 font-normal">(Future feature - disabled)</span>
              </label>
              <input
                type="text"
                id="product_title"
                value={productTitle}
                onChange={(e) => setProductTitle(e.target.value)}
                placeholder="Enter product title (optional override)"
                disabled
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm bg-gray-100 text-gray-500 cursor-not-allowed"
              />
              <p className="mt-1 text-xs text-gray-500">
                This field will allow manual override of product title in the future
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label htmlFor="lazada_url" className="block text-sm font-medium text-gray-700 mb-2">
                  Lazada Product URL <span className="text-red-500">*</span>
                </label>
                <input
                  type="url"
                  id="lazada_url"
                  value={lazadaUrl}
                  onChange={handleLazadaUrlChange}
                  placeholder="https://www.lazada.co.th/products/..."
                  className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white ${
                    urlErrors.lazada ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : 'border-gray-300'
                  }`}
                />
                {urlErrors.lazada && (
                  <p className="mt-1 text-xs text-red-600">{urlErrors.lazada}</p>
                )}
                <p className="mt-1 text-xs text-gray-500">
                  Enter Lazada product URL (optional if Shopee URL is provided, or provide both)
                </p>
              </div>
              <div>
                <label htmlFor="shopee_url" className="block text-sm font-medium text-gray-700 mb-2">
                  Shopee Product URL <span className="text-red-500">*</span>
                </label>
                <input
                  type="url"
                  id="shopee_url"
                  value={shopeeUrl}
                  onChange={handleShopeeUrlChange}
                  placeholder="https://shopee.co.th/product/..."
                  className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white ${
                    urlErrors.shopee ? 'border-red-300 focus:border-red-500 focus:ring-red-500' : 'border-gray-300'
                  }`}
                />
                {urlErrors.shopee && (
                  <p className="mt-1 text-xs text-red-600">{urlErrors.shopee}</p>
                )}
                <p className="mt-1 text-xs text-gray-500">
                  Enter Shopee product URL (optional if Lazada URL is provided, or provide both)
                </p>
              </div>
            </div>
            <p className="text-xs text-gray-500 -mt-2">
              <span className="text-red-500">*</span> At least one URL is required. You can provide both URLs to fetch offers from both marketplaces.
            </p>
            {error && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
                {error}
              </div>
            )}
            <button
              type="submit"
              disabled={loading}
              className="w-full sm:w-auto px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Adding...' : 'Add Product'}
            </button>
          </form>
        </div>

        {/* Product List */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900">Product List</h2>
          {loadingProducts ? (
            <p className="text-gray-500">Loading products...</p>
          ) : products.length === 0 ? (
            <p className="text-gray-500">No products yet. Add a product to get started.</p>
          ) : (
            <div className="space-y-4">
              {products.map((product) => {
                // Calculate best price
                const bestPrice = product.offers && product.offers.length > 0
                  ? product.offers.reduce((best, offer) => 
                      !best || offer.price < best.price ? offer : best
                    )
                  : null

                return (
                  <div
                    key={product.id}
                    className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50"
                  >
                    <div className="flex items-start space-x-4">
                      <img
                        src={product.image_url || '/placeholder-product.png'}
                        alt={product.title}
                        className="w-24 h-24 object-cover rounded"
                        onError={(e) => {
                          const target = e.target as HTMLImageElement;
                          target.src = '/placeholder-product.png';
                        }}
                      />
                      <div className="flex-1">
                        <h3 className="text-lg font-medium text-gray-900">{product.title}</h3>
                        <p className="text-sm text-gray-500">ID: {product.id}</p>
                        
                        {/* Offers */}
                        {product.offers && product.offers.length > 0 ? (
                          <div className="mt-3 space-y-2">
                            {bestPrice && (
                              <div className="bg-green-50 border border-green-200 rounded-lg p-2">
                                <p className="text-sm font-medium text-green-800">
                                  Best Price: {bestPrice.marketplace.toUpperCase()} - ฿{bestPrice.price.toFixed(2)}
                                </p>
                              </div>
                            )}
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                              {product.offers.map((offer) => {
                                const isBestPrice = bestPrice && bestPrice.marketplace === offer.marketplace && bestPrice.price === offer.price
                                return (
                                  <div
                                    key={offer.id}
                                    className={`border rounded-lg p-2 ${
                                      isBestPrice
                                        ? 'border-green-500 bg-green-50'
                                        : 'border-gray-200'
                                    }`}
                                  >
                                    <div className="flex justify-between items-start">
                                      <div>
                                        <p className="font-medium text-gray-900 capitalize text-sm">
                                          {offer.marketplace}
                                        </p>
                                        <p className="text-xs text-gray-600">{offer.store_name}</p>
                                      </div>
                                      <div className="text-right">
                                        <p className="text-base font-bold text-gray-900">
                                          ฿{offer.price.toFixed(2)}
                                        </p>
                                        {isBestPrice && (
                                          <span className="text-xs bg-green-500 text-white px-1 py-0.5 rounded">
                                            Best
                                          </span>
                                        )}
                                      </div>
                                    </div>
                                  </div>
                                )
                              })}
                            </div>
                          </div>
                        ) : (
                          <p className="text-sm text-gray-500 mt-2">No offers available</p>
                        )}

                        <div className="mt-3">
                          <button
                            onClick={() => handleDeleteProduct(product.id)}
                            disabled={deletingProductId === product.id}
                            className="text-sm text-red-600 hover:text-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
                          >
                            {deletingProductId === product.id ? 'Deleting...' : 'Delete'}
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>

      </div>
    </AdminLayout>
  )
}
