import type { Metadata } from "next";
import { Manrope } from "next/font/google";
import { Header } from "@/components/header";
import { Footer } from "@/components/footer";
import "./globals.css";

const manrope = Manrope({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Verinode — Institutional GPU Forward Marketplace & Telemetry Verification",
  description:
    "Institutional brokered marketplace for physically delivered, enterprise GPU-capacity reservations with cryptographic host telemetry verification.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={`dark ${manrope.variable}`}>
      <body className="min-h-screen bg-background text-foreground antialiased flex flex-col font-sans selection:bg-zinc-800 selection:text-white">
        <Header />
        <main className="flex-1">{children}</main>
        <Footer />
      </body>
    </html>
  );
}
