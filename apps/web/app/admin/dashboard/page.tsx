'use client'

import { useState, useEffect } from 'react'
import AdminLayout from '@/components/AdminLayout'

interface DashboardStats {
  total_clicks: number
  total_links: number
  ctr: number
  campaign_stats: Array<{
    campaign_id: string
    campaign_name: string
    clicks: number
  }>
  marketplace_stats: Array<{
    marketplace: string
    clicks: number
    percentage: number
  }>
  top_products: Array<{
    product_id: string
    product_name: string
    clicks: number
    marketplace: string
  }>
  recent_clicks: Array<{
    datetime: string
    product_id: string
    product_name: string
    marketplace: string
    campaign_id: string
    campaign_name: string
  }>
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'

async function fetchDashboardStats(marketplace?: string): Promise<DashboardStats> {
  const params = new URLSearchParams()
  if (marketplace) params.append('marketplace', marketplace)

  const response = await fetch(`${API_BASE_URL}/api/dashboard?${params.toString()}`)

  if (!response.ok) {
    throw new Error('Failed to fetch dashboard stats')
  }

  return response.json()
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [marketplaceFilter, setMarketplaceFilter] = useState('')

  useEffect(() => {
    loadStats()
  }, [marketplaceFilter])

  const loadStats = async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await fetchDashboardStats(
        marketplaceFilter || undefined
      )
      setStats(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard stats')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <AdminLayout>
        <div className="px-4 sm:px-0">
          <p className="text-gray-500">Loading dashboard...</p>
        </div>
      </AdminLayout>
    )
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="px-4 sm:px-0">
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        </div>
      </AdminLayout>
    )
  }

  if (!stats) {
    return null
  }

  return (
    <AdminLayout>
      <div className="px-4 sm:px-0">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">Dashboard</h1>

        {/* Filters */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Filters</h2>
          <div className="max-w-xs">
            <label htmlFor="marketplace" className="block text-sm font-medium text-gray-700 mb-2">
              Marketplace (optional)
            </label>
            <select
              id="marketplace"
              value={marketplaceFilter}
              onChange={(e) => setMarketplaceFilter(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
            >
              <option value="">All Marketplaces</option>
              <option value="lazada">Lazada</option>
              <option value="shopee">Shopee</option>
            </select>
          </div>
        </div>

        {/* Summary Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Clicks</h3>
            <p className="text-3xl font-bold text-gray-900">{stats.total_clicks.toLocaleString()}</p>
          </div>
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Links</h3>
            <p className="text-3xl font-bold text-gray-900">{stats.total_links.toLocaleString()}</p>
          </div>
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">CTR (Click-Through Rate)</h3>
            <p className="text-3xl font-bold text-gray-900">{stats.ctr.toFixed(2)}%</p>
          </div>
        </div>

        {/* Campaign Stats */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Campaign Statistics</h2>
          {stats.campaign_stats.length === 0 ? (
            <p className="text-gray-500">No campaign statistics available.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Campaign Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Clicks
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {stats.campaign_stats.map((campaign) => (
                    <tr key={campaign.campaign_id}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {campaign.campaign_name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {campaign.clicks.toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Marketplace Stats */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Marketplace Statistics</h2>
          {stats.marketplace_stats.length === 0 ? (
            <p className="text-gray-500">No marketplace statistics available.</p>
          ) : (
            <div className="space-y-4">
              {stats.marketplace_stats.map((marketplace) => (
                <div key={marketplace.marketplace} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-lg font-medium text-gray-900 capitalize">
                      {marketplace.marketplace}
                    </span>
                    <span className="text-lg font-bold text-gray-900">
                      {marketplace.clicks.toLocaleString()} clicks
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-primary-600 h-2 rounded-full"
                      style={{ width: `${marketplace.percentage}%` }}
                    />
                  </div>
                  <p className="text-sm text-gray-500 mt-1">{marketplace.percentage.toFixed(1)}%</p>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Top Products */}
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Top Performing Products</h2>
          {stats.top_products.length === 0 ? (
            <p className="text-gray-500">No top products available.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Product Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Marketplace
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Clicks
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {stats.top_products.map((product) => (
                    <tr key={`${product.product_id}-${product.marketplace}`}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {product.product_name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 capitalize">
                        {product.marketplace}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {product.clicks.toLocaleString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Click Report */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Click Report</h2>
          {stats.recent_clicks && stats.recent_clicks.length === 0 ? (
            <p className="text-gray-500">No recent clicks available.</p>
          ) : stats.recent_clicks && stats.recent_clicks.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      DateTime
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Product
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Marketplace
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Campaign
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {stats.recent_clicks.map((click, index) => (
                    <tr key={`${click.product_id}-${click.campaign_id}-${index}`}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {new Date(click.datetime).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {click.product_name}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 capitalize">
                        {click.marketplace}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {click.campaign_name}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-gray-500">No recent clicks available.</p>
          )}
        </div>
      </div>
    </AdminLayout>
  )
}
