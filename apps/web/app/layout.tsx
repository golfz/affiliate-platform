import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Jenosize Affiliate Platform',
  description: 'Affiliate link generation and price comparison platform',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
