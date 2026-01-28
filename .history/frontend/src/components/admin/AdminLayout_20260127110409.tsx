// src/components/admin/AdminLayout.tsx
import { useEffect, useState, type ReactNode } from 'react';

interface AdminUser {
	id: string;
	email: string;
	nombre_completo: string;
	rol: 'ADMIN' | 'SOPORTE';
}

interface AdminLayoutProps {
	children: ReactNode;
	currentPage: 'dashboard' | 'reclamos' | 'perfil';
}

export default function AdminLayout({ children, currentPage }: AdminLayoutProps) {
	const [user, setUser] = useState<AdminUser | null>(null);
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

	useEffect(() => {
		const token = localStorage.getItem('admin_token');
		const userStr = localStorage.getItem('admin_user');

		if (!token || !userStr) {
			window.location.href = '/admin/login';
			return;
		}

		try {
			setUser(JSON.parse(userStr));
		} catch {
			localStorage.removeItem('admin_token');
			localStorage.removeItem('admin_user');
			window.location.href = '/admin/login';
		}
	}, []);

	const handleLogout = () => {
		if (confirm('¿Estás seguro que deseas cerrar sesión?')) {
			localStorage.removeItem('admin_token');
			localStorage.removeItem('admin_user');
			window.location.href = '/admin/login';
		}
	};

	if (!user) {
		return (
			<div className="min-h-screen flex items-center justify-center">
				<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			</div>
		);
	}

	return (
		<div className="min-h-screen bg-gray-50">
			{/* Top Navigation */}
			<nav className="bg-white border-b border-gray-200 fixed w-full z-30 top-0">
				<div className="px-4 sm:px-6 lg:px-8">
					<div className="flex justify-between h-16">
						<div className="flex items-center">
							{/* Mobile menu button */}
							<button
								onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
								className="lg:hidden inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
							>
								<svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
								</svg>
							</button>

							{/* Logo */}
							<div className="flex-shrink-0 flex items-center ml-4 lg:ml-0">
								<div className="flex items-center gap-3">
									<div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
										<svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
										</svg>
									</div>
									<div className="hidden sm:block">
										<h1 className="text-lg font-bold text-gray-900">Panel Admin</h1>
										<p className="text-xs text-gray-500">CODEPLEX</p>
									</div>
								</div>
							</div>
						</div>

						{/* User Menu */}
						<div className="flex items-center gap-4">
							{/* User Info */}
							<div className="hidden md:flex items-center gap-3">
								<div className="text-right">
									<p className="text-sm font-medium text-gray-900">{user.nombre_completo}</p>
									<p className="text-xs text-gray-500">
										<span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
											user.rol === 'ADMIN' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'
										}`}>
											{user.rol}
										</span>
									</p>
								</div>
								<div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-semibold">
									{user.nombre_completo.charAt(0).toUpperCase()}
								</div>
							</div>

							{/* Logout Button */}
							<button
								onClick={handleLogout}
								className="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition"
							>
								<svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
								</svg>
								<span className="hidden sm:inline">Cerrar Sesión</span>
							</button>
						</div>
					</div>
				</div>
			</nav>

			{/* Mobile Sidebar Overlay */}
			{mobileMenuOpen && (
				<div
					className="fixed inset-0 bg-gray-600 bg-opacity-75 z-20 lg:hidden"
					onClick={() => setMobileMenuOpen(false)}
				></div>
			)}

			{/* Sidebar */}
			<div className={`fixed inset-y-0 left-0 w-64 bg-white border-r border-gray-200 pt-16 transform transition-transform duration-200 ease-in-out z-20 ${
				mobileMenuOpen ? 'translate-x-0' : '-translate-x-full'
			} lg:translate-x-0`}>
				<div className="h-full px-4 py-6 overflow-y-auto">
					<nav className="space-y-1">
						<a
							href="/admin/dashboard"
							className={`flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-lg transition ${
								currentPage === 'dashboard'
									? 'bg-blue-50 text-blue-700'
									: 'text-gray-700 hover:bg-gray-50'
							}`}
						>
							<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
							</svg>
							Dashboard
						</a>

						<a
							href="/admin/reclamos"
							className={`flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-lg transition ${
								currentPage === 'reclamos'
									? 'bg-blue-50 text-blue-700'
									: 'text-gray-700 hover:bg-gray-50'
							}`}
						>
							<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
							</svg>
							Reclamos
						</a>

						<div className="pt-6 mt-6 border-t border-gray-200">
							<p className="px-4 text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">
								Cuenta
							</p>
							<a
								href="/"
								target="_blank"
								className="flex items-center gap-3 px-4 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50 rounded-lg transition"
							>
								<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
								</svg>
								Ver sitio público
							</a>
						</div>
					</nav>
				</div>
			</div>

			{/* Main Content */}
			<div className="lg:pl-64 pt-16">
				<main className="py-8 px-4 sm:px-6 lg:px-8">
					{children}
				</main>
			</div>
		</div>
	);
}