import type { Metadata } from 'next';
import './globals.css';
import { Header } from '../components/layout/Header';
import { Footer } from '../components/layout/Footer';

export const metadata: Metadata = {
  title: 'Soroban Fleet Registry',
  description: 'Discover and verify Soroban CAP-85 externally managed contract executable fleets.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <Header />
        <main style={{ flex: 1, padding: '2.5rem 0' }}>
          {children}
        </main>
        <Footer />
      </body>
    </html>
  );
}
