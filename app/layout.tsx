import './globals.css'
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'

const font = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
});

export const metadata: Metadata = {
  title: 'Real-Time chat using Next.js and socket.io',
  description: 'A real-time chat application using end to end encryption',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={font.variable}>{children}</body>
    </html>
  )
}
