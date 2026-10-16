import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Layout } from './Layout';

const renderLayout = () =>
  render(
    <MemoryRouter initialEntries={['/']}>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<div>Outlet content</div>} />
        </Route>
      </Routes>
    </MemoryRouter>
  );

describe('Layout', () => {
  it('renders the brand, navigation links, and the outlet', () => {
    renderLayout();
    expect(screen.getByText('NEO Bank')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Dashboard' })).toHaveAttribute(
      'href',
      '/'
    );
    expect(screen.getByRole('link', { name: 'Transfer' })).toHaveAttribute(
      'href',
      '/transfer'
    );
    expect(screen.getByRole('link', { name: 'Transactions' })).toHaveAttribute(
      'href',
      '/transactions'
    );
    expect(screen.getByText('Outlet content')).toBeInTheDocument();
  });
});
