'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import { getPublicCampaign, getRedirectUrl, type CampaignPublicResponse } from '@/lib/api'

export default function CampaignPage() {
  const params = useParams()
  const campaignId = params.id as string
  const [campaign, setCampaign] = useState<CampaignPublicResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [copiedLink, setCopiedLink] = useState<string | null>(null)

  useEffect(() => {
    loadCampaign()
  }, [campaignId])

  const loadCampaign = async () => {
    try {
      const data = await getPublicCampaign(campaignId)
      setCampaign(data)
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to load campaign'
      setError(errorMsg)
    } finally {
      setLoading(false)
    }
  }

  const handleBuyClick = (shortCode: string) => {
    const redirectUrl = getRedirectUrl(shortCode)
    window.open(redirectUrl, '_blank', 'noopener,noreferrer')
  }

  const handleCopyLink = async (fullUrl: string, marketplace: string) => {
    try {
      await navigator.clipboard.writeText(fullUrl)
      setCopiedLink(`${marketplace}-${fullUrl}`)
      // Reset copied state after 2 seconds
      setTimeout(() => {
        setCopiedLink(null)
      }, 2000)
    } catch (err) {
      console.error('Failed to copy link:', err)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Loading campaign...</p>
      </div>
    )
  }

  if (error || !campaign) {
    // Check if error is about campaign not being active
    const isNotActiveError = error && error.toLowerCase().includes('not currently active')
    
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {isNotActiveError ? 'Campaign Not Active' : 'Campaign Not Found'}
          </h1>
          <p className="text-gray-600">{error || 'The campaign you are looking for does not exist.'}</p>
        </div>
      </div>
    )
  }

  const isActive = new Date(campaign.start_at) <= new Date() && new Date() <= new Date(campaign.end_at)

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">{campaign.name}</h1>
          <p className="text-gray-600">
            {new Date(campaign.start_at).toLocaleDateString()} - {new Date(campaign.end_at).toLocaleDateString()}
          </p>
          {!isActive && (
            <p className="mt-2 text-sm text-red-600">This campaign is not currently active.</p>
          )}
        </div>

        {campaign.products.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">No products available in this campaign.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {campaign.products.map((product) => (
              <div key={product.id} className="bg-white rounded-lg shadow-md overflow-hidden">
                <img
                  src={product.image_url || '/placeholder-product.png'}
                  alt={product.title}
                  className="w-full h-48 object-cover"
                  onError={(e) => {
                    const target = e.target as HTMLImageElement;
                    target.src = '/placeholder-product.png';
                  }}
                />
                <div className="p-6">
                  <h3 className="text-xl font-semibold text-gray-900 mb-4">{product.title}</h3>

                  {/* Offers */}
                  {product.offers.length > 0 && (
                    <div className="space-y-3 mb-4">
                      {product.offers.map((offer) => {
                        const isBestPrice = product.best_price?.marketplace === offer.marketplace
                        return (
                          <div
                            key={offer.id}
                            className={`border rounded-lg p-3 ${
                              isBestPrice ? 'border-green-500 bg-green-50' : 'border-gray-200'
                            }`}
                          >
                            <div className="flex justify-between items-center">
                              <div>
                                <p className="font-medium text-gray-900 capitalize">
                                  {offer.marketplace}
                                </p>
                                <p className="text-sm text-gray-600">{offer.store_name}</p>
                              </div>
                              <div className="text-right">
                                <p className="text-lg font-bold text-gray-900">
                                  ฿{offer.price.toFixed(2)}
                                </p>
                                {isBestPrice && (
                                  <span className="text-xs bg-green-500 text-white px-2 py-1 rounded">
                                    Best Price
                                  </span>
                                )}
                              </div>
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  )}

                  {/* Buy Buttons and Copy Links */}
                  <div className="space-y-2">
                    {product.offers.map((offer) => {
                      // Find the link for this marketplace - ensure case-insensitive matching
                      const link = product.links?.find(l => 
                        l.marketplace.toLowerCase() === offer.marketplace.toLowerCase()
                      )
                      if (!link) {
                        return null // No link available for this marketplace
                      }
                      const isCopied = copiedLink === `${offer.marketplace}-${link.full_url}`
                      const marketplaceName = offer.marketplace.charAt(0).toUpperCase() + offer.marketplace.slice(1)
                      return (
                        <div key={`${offer.id}-${offer.marketplace}`} className="flex items-center gap-2">
                          <button
                            onClick={() => handleBuyClick(link.short_code)}
                            className={`flex-1 px-4 py-2 rounded-lg font-medium transition ${
                              product.best_price?.marketplace === offer.marketplace
                                ? 'bg-green-600 text-white hover:bg-green-700'
                                : 'bg-primary-600 text-white hover:bg-primary-700'
                            }`}
                          >
                            Buy on {marketplaceName}
                          </button>
                          <div className="relative group">
                            <button
                              onClick={() => handleCopyLink(link.full_url, offer.marketplace)}
                              className={`px-3 py-2 rounded-lg font-medium transition border ${
                                isCopied
                                  ? 'bg-green-50 border-green-500 text-green-700'
                                  : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                              }`}
                              title={isCopied ? `Copied! ${marketplaceName} link copied to clipboard` : `Copy ${marketplaceName} link`}
                            >
                              {isCopied ? (
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                </svg>
                              ) : (
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                </svg>
                              )}
                            </button>
                            {/* Tooltip on hover */}
                            <div className="absolute bottom-full right-0 mb-2 px-3 py-1.5 bg-gray-900 text-white text-sm rounded-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-10">
                              {isCopied ? `Copied! ${marketplaceName} link copied` : `Copy ${marketplaceName} link`}
                              <div className="absolute top-full right-4 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-900"></div>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
