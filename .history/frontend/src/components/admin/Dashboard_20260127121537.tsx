// src/components/admin/Dashboard.tsx
import { useEffect, useState } from 'react';
import AdminLayout from './AdminLayout';

// Aseguramos que no haya slash al final para evitar dobles //
const RAW_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:3000';
const API_URL = RAW_URL.endsWith('/') ? RAW_URL.slice(0, -1) : RAW_URL;

interface Stats {
    // ... resto de tu interfaz igual
	total_reclamos: number;
	pendientes: number;
	en_proceso: number;
	resueltos: number;
	cerrados: number;
	reclamos_hoy: number;
	reclamos_semana: number;
	reclamos_mes: number;
	promedio_dias_resolucion: number | null;
}

export default function Dashboard() {
	const [stats, setStats] = useState<Stats | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		fetchStats();
	}, []);

	const fetchStats = async () => {
		try {
			const token = localStorage.getItem('admin_token');
			if (!token) {
				window.location.href = '/admin/login';
				return;
			}

			const response = await fetch(`${API_URL}/api/admin/dashboard/stats`, {
				headers: {
					'Authorization': `Bearer ${token}`,
				},
			});

			if (response.status === 401) {
				localStorage.removeItem('admin_token');
				localStorage.removeItem('admin_user');
				window.location.href = '/admin/login';
				return;
			}

			const data = await response.json();
			if (data.success) {
				setStats(data.data);
			} else {
				setError(data.message || 'Error al cargar estadísticas');
			}
		} catch (err: any) {
			setError(err.message || 'Error de conexión');
		} finally {
			setLoading(false);
		}
	};

	return (
		<AdminLayout currentPage="dashboard">
			<div className="space-y-6">
				{/* Header */}
				<div>
					<h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
					<p className="mt-2 text-gray-600">Resumen general del sistema de reclamos</p>
				</div>

				{loading && (
					<div className="flex items-center justify-center py-12">
						<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
					</div>
				)}

				{error && (
					<div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
						{error}
					</div>
				)}

				{stats && (
					<>
						{/* Stats Cards */}
						<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
							{/* Total Reclamos */}
							<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
								<div className="flex items-center justify-between">
									<div>
										<p className="text-sm font-medium text-gray-600">Total Reclamos</p>
										<p className="text-3xl font-bold text-gray-900 mt-2">{stats.total_reclamos}</p>
									</div>
									<div className="p-3 bg-blue-100 rounded-lg">
										<svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
										</svg>
									</div>
								</div>
							</div>

							{/* Pendientes */}
							<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
								<div className="flex items-center justify-between">
									<div>
										<p className="text-sm font-medium text-gray-600">Pendientes</p>
										<p className="text-3xl font-bold text-yellow-600 mt-2">{stats.pendientes}</p>
									</div>
									<div className="p-3 bg-yellow-100 rounded-lg">
										<svg className="w-6 h-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</div>
								</div>
								<div className="mt-4">
									<a href="/admin/reclamos?estado=PENDIENTE" className="text-sm text-yellow-600 hover:text-yellow-700 font-medium">
										Ver todos →
									</a>
								</div>
							</div>

							{/* En Proceso */}
							<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
								<div className="flex items-center justify-between">
									<div>
										<p className="text-sm font-medium text-gray-600">En Proceso</p>
										<p className="text-3xl font-bold text-blue-600 mt-2">{stats.en_proceso}</p>
									</div>
									<div className="p-3 bg-blue-100 rounded-lg">
										<svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
										</svg>
									</div>
								</div>
								<div className="mt-4">
									<a href="/admin/reclamos?estado=EN_PROCESO" className="text-sm text-blue-600 hover:text-blue-700 font-medium">
										Ver todos →
									</a>
								</div>
							</div>

							{/* Resueltos */}
							<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
								<div className="flex items-center justify-between">
									<div>
										<p className="text-sm font-medium text-gray-600">Resueltos</p>
										<p className="text-3xl font-bold text-green-600 mt-2">{stats.resueltos}</p>
									</div>
									<div className="p-3 bg-green-100 rounded-lg">
										<svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</div>
								</div>
								<div className="mt-4">
									<a href="/admin/reclamos?estado=RESUELTO" className="text-sm text-green-600 hover:text-green-700 font-medium">
										Ver todos →
									</a>
								</div>
							</div>
						</div>

						{/* Recent Activity */}
						<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
							{/* Actividad Reciente */}
							<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
								<h2 className="text-lg font-semibold text-gray-900 mb-4">Actividad Reciente</h2>
								<div className="space-y-4">
									<div className="flex items-center gap-4">
										<div className="p-2 bg-green-100 rounded-lg">
											<svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
											</svg>
										</div>
										<div className="flex-1">
											<p className="text-sm font-medium text-gray-900">Reclamos de hoy</p>
											<p className="text-sm text-gray-500">{stats.reclamos_hoy} nuevos</p>
										</div>
										<span className="text-xl font-bold text-gray-900">{stats.reclamos_hoy}</span>
									</div>

									<div className="flex items-center gap-4">
										<div className="p-2 bg-blue-100 rounded-lg">
											<svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
											</svg>
										</div>
										<div className="flex-1">
											<p className="text-sm font-medium text-gray-900">Esta semana</p>
											<p className="text-sm text-gray-500">Últimos 7 días</p>
										</div>
										<span className="text-xl font-bold text-gray-900">{stats.reclamos_semana}</span>
									</div>

									<div className="flex items-center gap-4">
										<div className="p-2 bg-purple-100 rounded-lg">
											<svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
											</svg>
										</div>
										<div className="flex-1">
											<p className="text-sm font-medium text-gray-900">Este mes</p>
											<p className="text-sm text-gray-500">Últimos 30 días</p>
										</div>
										<span className="text-xl font-bold text-gray-900">{stats.reclamos_mes}</span>
									</div>
								</div>
							</div>

							{/* Performance */}
							<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
								<h2 className="text-lg font-semibold text-gray-900 mb-4">Rendimiento</h2>
								<div className="space-y-6">
									{/* Promedio de Resolución */}
									<div>
										<div className="flex items-center justify-between mb-2">
											<span className="text-sm font-medium text-gray-700">Tiempo promedio de resolución</span>
											<span className="text-2xl font-bold text-blue-600">
												{stats.promedio_dias_resolucion ? Math.round(stats.promedio_dias_resolucion) : 0} días
											</span>
										</div>
										<div className="w-full bg-gray-200 rounded-full h-2">
											<div 
												className="bg-blue-600 h-2 rounded-full transition-all duration-500"
												style={{ 
													width: stats.promedio_dias_resolucion 
														? `${Math.min((stats.promedio_dias_resolucion / 15) * 100, 100)}%`
														: '0%'
												}}
											></div>
										</div>
										<p className="text-xs text-gray-500 mt-1">Límite legal: 15 días hábiles</p>
									</div>

									{/* Tasa de Resolución */}
									<div>
										<div className="flex items-center justify-between mb-2">
											<span className="text-sm font-medium text-gray-700">Tasa de resolución</span>
											<span className="text-2xl font-bold text-green-600">
												{stats.total_reclamos > 0 
													? Math.round(((stats.resueltos + stats.cerrados) / stats.total_reclamos) * 100)
													: 0}%
											</span>
										</div>
										<div className="w-full bg-gray-200 rounded-full h-2">
											<div 
												className="bg-green-600 h-2 rounded-full transition-all duration-500"
												style={{ 
													width: stats.total_reclamos > 0
														? `${((stats.resueltos + stats.cerrados) / stats.total_reclamos) * 100}%`
														: '0%'
												}}
											></div>
										</div>
										<p className="text-xs text-gray-500 mt-1">Reclamos resueltos y cerrados</p>
									</div>

									{/* Quick Actions */}
									<div className="pt-4 border-t border-gray-100">
										<p className="text-sm font-medium text-gray-700 mb-3">Acciones rápidas</p>
										<div className="grid grid-cols-2 gap-2">
											<a 
												href="/admin/reclamos?estado=PENDIENTE" 
												className="text-center py-2 px-3 bg-yellow-50 hover:bg-yellow-100 text-yellow-700 rounded-lg text-sm font-medium transition"
											>
												Ver Pendientes
											</a>
											<a 
												href="/admin/reclamos" 
												className="text-center py-2 px-3 bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-lg text-sm font-medium transition"
											>
												Todos los Reclamos
											</a>
										</div>
									</div>
								</div>
							</div>
						</div>

						{/* Status Distribution */}
						<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
							<h2 className="text-lg font-semibold text-gray-900 mb-6">Distribución por Estado</h2>
							<div className="grid grid-cols-2 md:grid-cols-5 gap-4">
								<div className="text-center">
									<div className="text-3xl font-bold text-yellow-600">{stats.pendientes}</div>
									<div className="text-sm text-gray-600 mt-1">Pendientes</div>
									<div className="text-xs text-gray-500 mt-0.5">
										{stats.total_reclamos > 0 ? Math.round((stats.pendientes / stats.total_reclamos) * 100) : 0}%
									</div>
								</div>
								<div className="text-center">
									<div className="text-3xl font-bold text-blue-600">{stats.en_proceso}</div>
									<div className="text-sm text-gray-600 mt-1">En Proceso</div>
									<div className="text-xs text-gray-500 mt-0.5">
										{stats.total_reclamos > 0 ? Math.round((stats.en_proceso / stats.total_reclamos) * 100) : 0}%
									</div>
								</div>
								<div className="text-center">
									<div className="text-3xl font-bold text-green-600">{stats.resueltos}</div>
									<div className="text-sm text-gray-600 mt-1">Resueltos</div>
									<div className="text-xs text-gray-500 mt-0.5">
										{stats.total_reclamos > 0 ? Math.round((stats.resueltos / stats.total_reclamos) * 100) : 0}%
									</div>
								</div>
								<div className="text-center">
									<div className="text-3xl font-bold text-gray-600">{stats.cerrados}</div>
									<div className="text-sm text-gray-600 mt-1">Cerrados</div>
									<div className="text-xs text-gray-500 mt-0.5">
										{stats.total_reclamos > 0 ? Math.round((stats.cerrados / stats.total_reclamos) * 100) : 0}%
									</div>
								</div>
								<div className="text-center">
									<div className="text-3xl font-bold text-blue-900">{stats.total_reclamos}</div>
									<div className="text-sm text-gray-600 mt-1">Total</div>
									<div className="text-xs text-gray-500 mt-0.5">100%</div>
								</div>
							</div>
						</div>
					</>
				)}
			</div>
		</AdminLayout>
	);
}