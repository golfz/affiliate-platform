import Link from 'next/link'

export default function Home() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Jenosize Affiliate Platform
        </h1>
        <p className="text-gray-600 mb-8">
          Affiliate link generation and price comparison platform
        </p>
        <div className="space-x-4">
          <Link
            href="/admin/products"
            className="inline-block px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition"
          >
            Admin Dashboard
          </Link>
        </div>
      </div>
    </div>
  )
}
