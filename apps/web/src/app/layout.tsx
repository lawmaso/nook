import type { Metadata } from "next"
import { Geist, Geist_Mono, Inter } from "next/font/google"

import { Providers } from "@/components/providers"
import { SiteHeader } from "@/components/site-header"

import "./globals.css"
import { cn } from "@/lib/utils";

const inter = Inter({subsets:['latin'],variable:'--font-sans'})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
})

export const metadata: Metadata = {
  title: { default: "Yagami", template: "%s · Yagami" },
  description: "Leaderboards, head-to-head comparisons and trends from the Yagami scraper.",
}

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={cn("h-full", "antialiased", geistMono.variable, "font-sans", inter.variable)}
    >
      <body className="flex min-h-full flex-col bg-background text-foreground">
        <Providers>
          <SiteHeader />
          <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-6">{children}</main>
        </Providers>
      </body>
    </html>
  )
}
