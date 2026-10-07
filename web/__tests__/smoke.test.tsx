import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import HomePage from '../app/page';

describe('HomePage', () => {
  it('renders the title and Browse Fleets action', () => {
    render(<HomePage />);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Soroban Fleet Registry');
    expect(screen.getByText('Discover and verify CAP-85 executable fleets.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Browse fleets' })).toHaveAttribute('href', '/fleets');
  });
});
