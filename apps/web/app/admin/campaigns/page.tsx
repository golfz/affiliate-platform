'use client'

import { useState, useEffect } from 'react'
import AdminLayout from '@/components/AdminLayout'
import { getAllCampaigns, getCampaign, createCampaign, deleteCampaign, getAllProducts, updateCampaign, type CampaignResponse, type ProductResponse } from '@/lib/api'

export default function CampaignsPage() {
  const [name, setName] = useState('')
  const [utmCampaign, setUtmCampaign] = useState('')
  const [startAt, setStartAt] = useState('')
  const [endAt, setEndAt] = useState('')
  const [productIds, setProductIds] = useState<string[]>([])
  const [selectedProducts, setSelectedProducts] = useState<ProductResponse[]>([])
  const [showProductModal, setShowProductModal] = useState(false)
  const [availableProducts, setAvailableProducts] = useState<ProductResponse[]>([])
  const [loadingProducts, setLoadingProducts] = useState(false)
  const [modalSelectedIds, setModalSelectedIds] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [campaigns, setCampaigns] = useState<CampaignResponse[]>([])
  const [loadingCampaigns, setLoadingCampaigns] = useState(true)
  const [deletingCampaignId, setDeletingCampaignId] = useState<string | null>(null)
  const [editingCampaignId, setEditingCampaignId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [editUtmCampaign, setEditUtmCampaign] = useState('')
  const [editStartAt, setEditStartAt] = useState('')
  const [editEndAt, setEditEndAt] = useState('')
  const [editProductIds, setEditProductIds] = useState<string[]>([])
  const [editSelectedProducts, setEditSelectedProducts] = useState<ProductResponse[]>([])
  const [showEditCampaignModal, setShowEditCampaignModal] = useState(false)
  const [showEditProductModal, setShowEditProductModal] = useState(false)
  const [editModalSelectedIds, setEditModalSelectedIds] = useState<Set<string>>(new Set())
  const [updatingCampaign, setUpdatingCampaign] = useState(false)

  // Fetch campaigns on mount
  useEffect(() => {
    const fetchCampaigns = async () => {
      try {
        setLoadingCampaigns(true)
        const campaignsList = await getAllCampaigns()
        setCampaigns(campaignsList)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch campaigns')
      } finally {
        setLoadingCampaigns(false)
      }
    }

    fetchCampaigns()
  }, [])

  const handleOpenProductModal = async () => {
    setShowProductModal(true)
    setLoadingProducts(true)
    try {
      const products = await getAllProducts()
      setAvailableProducts(products)
      // Initialize modal selection with currently selected products
      setModalSelectedIds(new Set(productIds))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load products')
    } finally {
      setLoadingProducts(false)
    }
  }

  const handleToggleProductSelection = (productId: string) => {
    setModalSelectedIds(prev => {
      const newSet = new Set(prev)
      if (newSet.has(productId)) {
        newSet.delete(productId)
      } else {
        newSet.add(productId)
      }
      return newSet
    })
  }

  const handleConfirmProductSelection = () => {
    const selected = availableProducts.filter(p => modalSelectedIds.has(p.id))
    setSelectedProducts(selected)
    setProductIds(selected.map(p => p.id))
    setShowProductModal(false)
  }

  const handleRemoveProduct = (productId: string) => {
    setSelectedProducts(selectedProducts.filter(p => p.id !== productId))
    setProductIds(productIds.filter(id => id !== productId))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      // Convert datetime-local format (YYYY-MM-DDTHH:mm) to ISO 8601 UTC
      // datetime-local is in user's local timezone, we need to convert to UTC
      const convertToISO = (datetimeLocal: string): string => {
        if (!datetimeLocal) return ''
        // Create a Date object from the datetime-local string (interpreted as local time)
        const localDate = new Date(datetimeLocal)
        // Convert to ISO string and extract UTC part (replace timezone offset with Z)
        // Example: "2026-01-17T17:31:00+07:00" -> "2026-01-17T10:31:00Z"
        return localDate.toISOString()
      }

      await createCampaign({
        name,
        utm_campaign: utmCampaign,
        start_at: convertToISO(startAt),
        end_at: convertToISO(endAt),
        product_ids: productIds.length > 0 ? productIds : undefined,
      })
      
      // Refresh campaigns list
      const campaignsList = await getAllCampaigns()
      setCampaigns(campaignsList)

      // Clear form after successful creation
      setName('')
      setUtmCampaign('')
      setStartAt('')
      setEndAt('')
      setProductIds([])
      setSelectedProducts([])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create campaign')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteCampaign = async (campaignId: string) => {
    if (!confirm('Are you sure you want to delete this campaign? This will also delete all related links, campaign products, and clicks.')) {
      return
    }

    try {
      setDeletingCampaignId(campaignId)
      setError(null)
      await deleteCampaign(campaignId)
      // Remove from list
      setCampaigns(campaigns.filter(c => c.id !== campaignId))
      // Clear editing state if it was the campaign being edited
      if (editingCampaignId === campaignId) {
        setEditingCampaignId(null)
        setEditProductIds([])
        setEditSelectedProducts([])
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete campaign')
    } finally {
      setDeletingCampaignId(null)
    }
  }

  const handleEditCampaign = async (campaignId: string) => {
    setEditingCampaignId(campaignId)
    setLoadingProducts(true)
    setError(null)

    try {
      // Fetch campaign details including product IDs
      const campaign = await getCampaign(campaignId)
      
      setEditName(campaign.name)
      setEditUtmCampaign(campaign.utm_campaign)
      
      // Convert ISO dates to datetime-local format
      const startDate = new Date(campaign.start_at)
      const endDate = new Date(campaign.end_at)
      setEditStartAt(startDate.toISOString().slice(0, 16))
      setEditEndAt(endDate.toISOString().slice(0, 16))

      // Load all products
      const allProducts = await getAllProducts()
      setAvailableProducts(allProducts)
      
      // Set current campaign products
      const currentProductIds = campaign.product_ids || []
      setEditProductIds(currentProductIds)
      
      // Find and set selected products
      const selectedProducts = allProducts.filter(p => currentProductIds.includes(p.id))
      setEditSelectedProducts(selectedProducts)
      setEditModalSelectedIds(new Set(currentProductIds))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load campaign')
    } finally {
      setLoadingProducts(false)
    }

    setShowEditCampaignModal(true)
  }

  const handleCancelEdit = () => {
    setEditingCampaignId(null)
    setEditName('')
    setEditUtmCampaign('')
    setEditStartAt('')
    setEditEndAt('')
    setEditProductIds([])
    setEditSelectedProducts([])
    setShowEditCampaignModal(false)
    setShowEditProductModal(false)
  }

  const handleOpenEditProductModal = async () => {
    setShowEditProductModal(true)
    setLoadingProducts(true)
    try {
      const products = await getAllProducts()
      setAvailableProducts(products)
      setEditModalSelectedIds(new Set(editProductIds))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load products')
    } finally {
      setLoadingProducts(false)
    }
  }

  const handleConfirmEditProductSelection = () => {
    const selected = availableProducts.filter(p => editModalSelectedIds.has(p.id))
    setEditSelectedProducts(selected)
    setEditProductIds(selected.map(p => p.id))
    setShowEditProductModal(false)
  }

  const handleRemoveEditProduct = (productId: string) => {
    setEditSelectedProducts(editSelectedProducts.filter(p => p.id !== productId))
    setEditProductIds(editProductIds.filter(id => id !== productId))
  }

  const handleUpdateCampaign = async () => {
    if (!editingCampaignId) return

    setUpdatingCampaign(true)
    setError(null)

    try {
      // Convert datetime-local to ISO
      const convertToISO = (datetimeLocal: string): string => {
        if (!datetimeLocal) return ''
        const localDate = new Date(datetimeLocal)
        return localDate.toISOString()
      }

      await updateCampaign(editingCampaignId, {
        name: editName,
        utm_campaign: editUtmCampaign,
        start_at: convertToISO(editStartAt),
        end_at: convertToISO(editEndAt),
        product_ids: editProductIds.length > 0 ? editProductIds : undefined,
      })
      
      // Refresh campaigns list
      const campaignsList = await getAllCampaigns()
      setCampaigns(campaignsList)

      // Clear editing state
      handleCancelEdit()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update campaign')
    } finally {
      setUpdatingCampaign(false)
    }
  }

  return (
    <AdminLayout>
      <div className="px-4 sm:px-0">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Campaigns</h1>

        {/* Create Campaign Form */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900">Create Campaign</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                Campaign Name
              </label>
              <input
                type="text"
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Summer Deal 2025"
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                required
              />
            </div>

            <div>
              <label htmlFor="utm_campaign" className="block text-sm font-medium text-gray-700 mb-2">
                UTM Campaign
              </label>
              <input
                type="text"
                id="utm_campaign"
                value={utmCampaign}
                onChange={(e) => setUtmCampaign(e.target.value)}
                placeholder="summer_2025"
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                required
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label htmlFor="start_at" className="block text-sm font-medium text-gray-700 mb-2">
                  Start Date
                </label>
                <input
                  type="datetime-local"
                  id="start_at"
                  value={startAt}
                  onChange={(e) => setStartAt(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                  required
                />
              </div>

              <div>
                <label htmlFor="end_at" className="block text-sm font-medium text-gray-700 mb-2">
                  End Date
                </label>
                <input
                  type="datetime-local"
                  id="end_at"
                  value={endAt}
                  onChange={(e) => setEndAt(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                  required
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Products <span className="text-gray-500 text-xs font-normal">(Optional - can be added later)</span>
              </label>
              <button
                type="button"
                onClick={handleOpenProductModal}
                className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 mb-2"
              >
                Browse Products
              </button>
              {selectedProducts.length > 0 && (
                <div className="mt-3 space-y-2">
                  <p className="text-sm text-gray-600">
                    {selectedProducts.length} product(s) selected
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {selectedProducts.map((product) => (
                      <span
                        key={product.id}
                        className="inline-flex items-center px-3 py-1 rounded-full text-sm bg-primary-100 text-primary-800"
                      >
                        {product.title}
                        <button
                          type="button"
                          onClick={() => handleRemoveProduct(product.id)}
                          className="ml-2 text-primary-600 hover:text-primary-800"
                        >
                          ×
                        </button>
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>

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
              {loading ? 'Creating...' : 'Create Campaign'}
            </button>
          </form>
        </div>

        {/* Campaign List */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900">Campaign List</h2>
          {loadingCampaigns ? (
            <p className="text-gray-500">Loading campaigns...</p>
          ) : campaigns.length === 0 ? (
            <p className="text-gray-500">No campaigns yet. Create a campaign to get started.</p>
          ) : (
            <div className="space-y-4">
              {campaigns.map((c) => (
                <div
                  key={c.id}
                  className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50"
                >
                    <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <h3 className="text-lg font-medium text-gray-900">{c.name}</h3>
                      <p className="text-sm text-gray-500">UTM: {c.utm_campaign}</p>
                      <p className="text-sm text-gray-500">
                        {new Date(c.start_at).toLocaleDateString()} - {new Date(c.end_at).toLocaleDateString()}
                      </p>
                      <p className="text-xs text-gray-400 mt-1">ID: {c.id}</p>
                    </div>
                    <div className="flex flex-col items-end space-y-2">
                      <a
                        href={`/campaign/${c.id}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-primary-600 hover:text-primary-700"
                      >
                        View Public Page
                      </a>
                      <button
                        onClick={() => handleEditCampaign(c.id)}
                        disabled={editingCampaignId === c.id || deletingCampaignId === c.id}
                        className="text-sm text-blue-600 hover:text-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Edit Campaign
                      </button>
                      <button
                        onClick={() => handleDeleteCampaign(c.id)}
                        disabled={deletingCampaignId === c.id}
                        className="text-sm text-red-600 hover:text-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {deletingCampaignId === c.id ? 'Deleting...' : 'Delete'}
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>


        {/* Product Selection Modal (for creating new campaign) */}
        {showProductModal && !editingCampaignId && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[80vh] flex flex-col">
              <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
                <h2 className="text-xl font-semibold text-gray-900">Select Products</h2>
                <button
                  onClick={() => setShowProductModal(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="px-6 py-4 overflow-y-auto flex-1">
                {loadingProducts ? (
                  <p className="text-gray-500">Loading products...</p>
                ) : availableProducts.length === 0 ? (
                  <p className="text-gray-500">No products available. Please add products first.</p>
                ) : (
                  <div className="space-y-3">
                    {availableProducts.map((product) => {
                      const isSelected = modalSelectedIds.has(product.id)
                      return (
                        <div
                          key={product.id}
                          onClick={() => handleToggleProductSelection(product.id)}
                          className={`border rounded-lg p-3 cursor-pointer hover:bg-gray-50 ${
                            isSelected ? 'border-primary-500 bg-primary-50' : 'border-gray-200'
                          }`}
                        >
                          <div className="flex items-center space-x-3">
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={() => handleToggleProductSelection(product.id)}
                              className="w-4 h-4 text-primary-600 border-gray-300 rounded focus:ring-primary-500"
                              onClick={(e) => e.stopPropagation()}
                            />
                            <img
                              src={product.image_url || '/placeholder-product.png'}
                              alt={product.title}
                              className="w-16 h-16 object-cover rounded"
                              onError={(e) => {
                                const target = e.target as HTMLImageElement;
                                target.src = '/placeholder-product.png';
                              }}
                            />
                            <div className="flex-1">
                              <h3 className="text-sm font-medium text-gray-900">{product.title}</h3>
                              <p className="text-xs text-gray-500">ID: {product.id}</p>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
              <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
                <button
                  onClick={() => {
                    setShowProductModal(false)
                    if (editingCampaignId) {
                      setShowEditProductModal(false)
                    }
                  }}
                  className="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300"
                >
                  Cancel
                </button>
                <button
                  onClick={() => {
                    if (editingCampaignId) {
                      handleConfirmEditProductSelection()
                    } else {
                      handleConfirmProductSelection()
                    }
                  }}
                  className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
                >
                  Select ({modalSelectedIds.size})
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Edit Campaign Modal */}
        {showEditCampaignModal && editingCampaignId && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] flex flex-col">
              <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
                <h2 className="text-xl font-semibold text-gray-900">Edit Campaign</h2>
                <button
                  onClick={handleCancelEdit}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="px-6 py-4 overflow-y-auto flex-1">
                <form onSubmit={(e) => { e.preventDefault(); handleUpdateCampaign(); }} className="space-y-4">
                  <div>
                    <label htmlFor="edit_name" className="block text-sm font-medium text-gray-700 mb-2">
                      Campaign Name
                    </label>
                    <input
                      type="text"
                      id="edit_name"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      placeholder="Summer Deal 2025"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                      required
                    />
                  </div>

                  <div>
                    <label htmlFor="edit_utm_campaign" className="block text-sm font-medium text-gray-700 mb-2">
                      UTM Campaign
                    </label>
                    <input
                      type="text"
                      id="edit_utm_campaign"
                      value={editUtmCampaign}
                      onChange={(e) => setEditUtmCampaign(e.target.value)}
                      placeholder="summer_2025"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                      required
                    />
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label htmlFor="edit_start_at" className="block text-sm font-medium text-gray-700 mb-2">
                        Start Date
                      </label>
                      <input
                        type="datetime-local"
                        id="edit_start_at"
                        value={editStartAt}
                        onChange={(e) => setEditStartAt(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                        required
                      />
                    </div>

                    <div>
                      <label htmlFor="edit_end_at" className="block text-sm font-medium text-gray-700 mb-2">
                        End Date
                      </label>
                      <input
                        type="datetime-local"
                        id="edit_end_at"
                        value={editEndAt}
                        onChange={(e) => setEditEndAt(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-gray-900 bg-white"
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Products <span className="text-gray-500 text-xs font-normal">(Optional)</span>
                    </label>
                    <button
                      type="button"
                      onClick={handleOpenEditProductModal}
                      className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 mb-2"
                    >
                      Browse Products
                    </button>
                    {editSelectedProducts.length > 0 && (
                      <div className="mt-3 space-y-2">
                        <p className="text-sm text-gray-600">
                          {editSelectedProducts.length} product(s) selected
                        </p>
                        <div className="flex flex-wrap gap-2">
                          {editSelectedProducts.map((product) => (
                            <span
                              key={product.id}
                              className="inline-flex items-center px-3 py-1 rounded-full text-sm bg-primary-100 text-primary-800"
                            >
                              {product.title}
                              <button
                                type="button"
                                onClick={() => handleRemoveEditProduct(product.id)}
                                className="ml-2 text-primary-600 hover:text-primary-800"
                              >
                                ×
                              </button>
                            </span>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>

                  {error && (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
                      {error}
                    </div>
                  )}

                  <div className="flex space-x-3 pt-4">
                    <button
                      type="submit"
                      disabled={updatingCampaign}
                      className="flex-1 px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {updatingCampaign ? 'Saving...' : 'Save Changes'}
                    </button>
                    <button
                      type="button"
                      onClick={handleCancelEdit}
                      disabled={updatingCampaign}
                      className="px-6 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Cancel
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        )}

        {/* Edit Product Selection Modal */}
        {showEditProductModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[80vh] flex flex-col">
              <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
                <h2 className="text-xl font-semibold text-gray-900">Select Products</h2>
                <button
                  onClick={() => {
                    setShowEditProductModal(false)
                  }}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="px-6 py-4 overflow-y-auto flex-1">
                {loadingProducts ? (
                  <p className="text-gray-500">Loading products...</p>
                ) : availableProducts.length === 0 ? (
                  <p className="text-gray-500">No products available. Please add products first.</p>
                ) : (
                  <div className="space-y-3">
                    {availableProducts.map((product) => {
                      const isSelected = editModalSelectedIds.has(product.id)
                      return (
                        <div
                          key={product.id}
                          onClick={() => {
                            setEditModalSelectedIds(prev => {
                              const newSet = new Set(prev)
                              if (newSet.has(product.id)) {
                                newSet.delete(product.id)
                              } else {
                                newSet.add(product.id)
                              }
                              return newSet
                            })
                          }}
                          className={`border rounded-lg p-3 cursor-pointer hover:bg-gray-50 ${
                            isSelected ? 'border-primary-500 bg-primary-50' : 'border-gray-200'
                          }`}
                        >
                          <div className="flex items-center space-x-3">
                            <input
                              type="checkbox"
                              checked={isSelected}
                              onChange={(e) => {
                                e.stopPropagation()
                                setEditModalSelectedIds(prev => {
                                  const newSet = new Set(prev)
                                  if (newSet.has(product.id)) {
                                    newSet.delete(product.id)
                                  } else {
                                    newSet.add(product.id)
                                  }
                                  return newSet
                                })
                              }}
                              onClick={(e) => {
                                e.stopPropagation()
                              }}
                              className="w-4 h-4 text-primary-600 border-gray-300 rounded focus:ring-primary-500 cursor-pointer"
                            />
                            <img
                              src={product.image_url || '/placeholder-product.png'}
                              alt={product.title}
                              className="w-16 h-16 object-cover rounded"
                              onError={(e) => {
                                const target = e.target as HTMLImageElement;
                                target.src = '/placeholder-product.png';
                              }}
                            />
                            <div className="flex-1">
                              <p className="font-medium text-gray-900">{product.title}</p>
                              <p className="text-xs text-gray-500">ID: {product.id}</p>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                )}
              </div>
              <div className="px-6 py-4 border-t border-gray-200 flex justify-end space-x-3">
                <button
                  onClick={() => {
                    setShowEditProductModal(false)
                  }}
                  className="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
                >
                  Cancel
                </button>
                <button
                  onClick={handleConfirmEditProductSelection}
                  className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
                >
                  Confirm
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </AdminLayout>
  )
}
