import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { AddressDisplay } from '../components/common/AddressDisplay';
import { HashDisplay } from '../components/common/HashDisplay';
import { FleetStatus } from '../components/fleet/FleetStatus';
import { LoadingState } from '../components/common/LoadingState';
import { ErrorState } from '../components/common/ErrorState';
import { EmptyState } from '../components/common/EmptyState';
import { ApiError } from '../lib/api/client';

describe('Layout and Common Components', () => {
  it('renders AddressDisplay and supports copying', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    const addr = 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB';
    render(<AddressDisplay address={addr} />);

    expect(screen.getByText('CA3D5KRY...2AAAAB')).toBeInTheDocument();
    const copyBtn = screen.getByRole('button', { name: 'Copy address' });
    fireEvent.click(copyBtn);

    expect(writeTextMock).toHaveBeenCalledWith(addr);
    await waitFor(() => {
      expect(screen.getByText('Copied')).toBeInTheDocument();
    });
  });

  it('renders HashDisplay normalizing base64 to hex', async () => {
    const b64 = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=';
    render(<HashDisplay hash={b64} />);

    expect(screen.getByText('00000000...000000')).toBeInTheDocument();
  });

  it('renders FleetStatus for all 5 states with text and symbols', () => {
    const { rerender } = render(<FleetStatus status="HEALTHY" showDescription />);
    expect(screen.getByText('HEALTHY')).toBeInTheDocument();
    expect(screen.getByText('✓')).toBeInTheDocument();
    expect(screen.getByText(/All indexed active members/)).toBeInTheDocument();

    rerender(<FleetStatus status="DRIFT" showDescription />);
    expect(screen.getByText('DRIFT')).toBeInTheDocument();
    expect(screen.getByText('⚠')).toBeInTheDocument();
    expect(screen.getByText(/One or more active members resolve to a different executable/)).toBeInTheDocument();

    rerender(<FleetStatus status="BROKEN_REFERENCE" />);
    expect(screen.getByText('BROKEN_REFERENCE')).toBeInTheDocument();
    expect(screen.getByText('✕')).toBeInTheDocument();

    rerender(<FleetStatus status="INCOMPLETE" />);
    expect(screen.getByText('INCOMPLETE')).toBeInTheDocument();
    expect(screen.getByText('⏳')).toBeInTheDocument();

    rerender(<FleetStatus status="UNKNOWN" />);
    expect(screen.getByText('UNKNOWN')).toBeInTheDocument();
    expect(screen.getByText('?')).toBeInTheDocument();
  });

  it('renders LoadingState with accessible status role', () => {
    render(<LoadingState label="Indexing in progress..." />);
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByText('Indexing in progress...')).toBeInTheDocument();
  });

  it('renders ErrorState with ApiError details and retry button', () => {
    const onRetry = vi.fn();
    const err = new ApiError(404, 'Fleet not found', 'FLEET_NOT_FOUND');

    render(<ErrorState error={err} onRetry={onRetry} />);

    expect(screen.getByText('Not Found')).toBeInTheDocument();
    expect(screen.getByText('Fleet not found')).toBeInTheDocument();
    expect(screen.getByText('Code: FLEET_NOT_FOUND')).toBeInTheDocument();

    const retryBtn = screen.getByRole('button', { name: 'Retry' });
    fireEvent.click(retryBtn);
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it('renders EmptyState with custom message and link', () => {
    render(
      <EmptyState
        title="No Fleets"
        message="No CAP-85 fleets have been indexed yet."
        actionLabel="Browse Fleets"
        actionHref="/fleets"
      />
    );

    expect(screen.getByText('No Fleets')).toBeInTheDocument();
    expect(screen.getByText('No CAP-85 fleets have been indexed yet.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Browse Fleets' })).toHaveAttribute('href', '/fleets');
  });
});
