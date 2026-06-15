import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Ting Boundless',
  description: 'Public site (Next.js SSR via Gateway)',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
