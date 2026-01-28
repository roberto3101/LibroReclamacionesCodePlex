// src/components/admin/ReclamosList.tsx
import { useEffect, useState } from 'react';
import AdminLayout from './AdminLayout';

const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:3000';

interface Reclamo {
	id: string;
	codigo_reclamo: string;
	tipo_solicitud: string;
	estado: string;
	nombre_completo: string;
	email: string;
	telefono: string;
	descripcion_bien: string;
	fecha_registro: string;
	fecha_limite_respuesta: string;
	dias_restantes: number | null;
}

export default function ReclamosList() {
	const [reclamos, setReclamos] = useState<Reclamo[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [filtroEstado, setFiltroEstado] = useState('');
	const [searchTerm, setSearchTerm] = useState('');

	useEffect(() => {
		fetchReclamos();
	}, [filtroEstado]);

	const fetchReclamos = async () => {
		try {
			setLoading(true);
			const token = localStorage.getItem('admin_token');
			if (!token) {
				window.location.href = '/admin/login';
				return;
			}

			const url = filtroEstado 
				? `${API_URL}/api/admin/reclamos?estado=${filtroEstado}`
				: `${API_URL}/api/admin/reclamos`;

			const response = await fetch(url, {
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
				setReclamos(data.data);
			} else {
				setError(data.message || 'Error al cargar reclamos');
			}
		} catch (err: any) {
			setError(err.message || 'Error de conexión');
		} finally {
			setLoading(false);
		}
	};

	const getEstadoBadgeClass = (estado: string) => {
		switch (estado) {
			case 'PENDIENTE':
				return 'bg-yellow-100 text-yellow-800';
			case 'EN_PROCESO':
				return 'bg-blue-100 text-blue-800';
			case 'RESUELTO':
				return 'bg-green-100 text-green-800';
			case 'CERRADO':
				return 'bg-gray-100 text-gray-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	};

	const getTipoIcon = (tipo: string) => {
		if (tipo === 'RECLAMO') {
			return (
				<svg className="w-5 h-5 text-red-500" fill="currentColor" viewBox="0 0 20 20">
					<path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
				</svg>
			);
		}
		return (
			<svg className="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
				<path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
			</svg>
		);
	};

	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString('es-PE', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	};

	const filteredReclamos = reclamos.filter(r => 
		r.codigo_reclamo.toLowerCase().includes(searchTerm.toLowerCase()) ||
		r.nombre_completo.toLowerCase().includes(searchTerm.toLowerCase()) ||
		r.email.toLowerCase().includes(searchTerm.toLowerCase())
	);

	return (
		<AdminLayout currentPage="reclamos">
			<div className="space-y-6">
				{/* Header */}
				<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
					<div>
						<h1 className="text-3xl font-bold text-gray-900">Reclamos</h1>
						<p className="mt-2 text-gray-600">
							Gestiona todos los reclamos y quejas del sistema
						</p>
					</div>
					<button
						onClick={fetchReclamos}
						className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition"
					>
						<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
						</svg>
						Actualizar
					</button>
				</div>

				{/* Filters */}
				<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
					<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
						{/* Search */}
						<div>
							<label className="block text-sm font-medium text-gray-700 mb-2">
								Buscar
							</label>
							<div className="relative">
								<div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
									<svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
									</svg>
								</div>
								<input
									type="text"
									value={searchTerm}
									onChange={(e) => setSearchTerm(e.target.value)}
									placeholder="Código, nombre o email..."
									className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								/>
							</div>
						</div>

						{/* Estado Filter */}
						<div>
							<label className="block text-sm font-medium text-gray-700 mb-2">
								Filtrar por Estado
							</label>
							<select
								value={filtroEstado}
								onChange={(e) => setFiltroEstado(e.target.value)}
								className="block w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							>
								<option value="">Todos los estados</option>
								<option value="PENDIENTE">Pendientes</option>
								<option value="EN_PROCESO">En Proceso</option>
								<option value="RESUELTO">Resueltos</option>
								<option value="CERRADO">Cerrados</option>
							</select>
						</div>
					</div>
				</div>

				{/* Stats Summary */}
				<div className="grid grid-cols-2 md:grid-cols-4 gap-4">
					<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
						<div className="text-sm text-gray-600">Total</div>
						<div className="text-2xl font-bold text-gray-900 mt-1">{reclamos.length}</div>
					</div>
					<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
						<div className="text-sm text-gray-600">Pendientes</div>
						<div className="text-2xl font-bold text-yellow-600 mt-1">
							{reclamos.filter(r => r.estado === 'PENDIENTE').length}
						</div>
					</div>
					<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
						<div className="text-sm text-gray-600">En Proceso</div>
						<div className="text-2xl font-bold text-blue-600 mt-1">
							{reclamos.filter(r => r.estado === 'EN_PROCESO').length}
						</div>
					</div>
					<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
						<div className="text-sm text-gray-600">Resueltos</div>
						<div className="text-2xl font-bold text-green-600 mt-1">
							{reclamos.filter(r => r.estado === 'RESUELTO').length}
						</div>
					</div>
				</div>

				{/* Loading */}
				{loading && (
					<div className="flex items-center justify-center py-12">
						<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
					</div>
				)}

				{/* Error */}
				{error && (
					<div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
						{error}
					</div>
				)}

				{/* Reclamos List */}
				{!loading && !error && (
					<div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
						{filteredReclamos.length === 0 ? (
							<div className="text-center py-12">
								<svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
								<h3 className="mt-2 text-sm font-medium text-gray-900">No hay reclamos</h3>
								<p className="mt-1 text-sm text-gray-500">
									{searchTerm ? 'No se encontraron resultados para tu búsqueda' : 'No hay reclamos con los filtros seleccionados'}
								</p>
							</div>
						) : (
							<div className="overflow-x-auto">
								<table className="min-w-full divide-y divide-gray-200">
									<thead className="bg-gray-50">
										<tr>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Código / Cliente
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Tipo
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Estado
											</th>
											<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
												Fecha / Plazo
											</th>
											<th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
												Acciones
											</th>
										</tr>
									</thead>
									<tbody className="bg-white divide-y divide-gray-200">
										{filteredReclamos.map((reclamo) => (
											<tr key={reclamo.id} className="hover:bg-gray-50 transition">
												<td className="px-6 py-4">
													<div className="flex items-center gap-3">
														{getTipoIcon(reclamo.tipo_solicitud)}
														<div>
															<div className="text-sm font-medium text-gray-900">
																{reclamo.codigo_reclamo}
															</div>
															<div className="text-sm text-gray-500">
																{reclamo.nombre_completo}
															</div>
															<div className="text-xs text-gray-400">
																{reclamo.email}
															</div>
														</div>
													</div>
												</td>
												<td className="px-6 py-4">
													<span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
														reclamo.tipo_solicitud === 'RECLAMO' 
															? 'bg-red-100 text-red-800'
															: 'bg-blue-100 text-blue-800'
													}`}>
														{reclamo.tipo_solicitud}
													</span>
												</td>
												<td className="px-6 py-4">
													<span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getEstadoBadgeClass(reclamo.estado)}`}>
														{reclamo.estado.replace('_', ' ')}
													</span>
												</td>
												<td className="px-6 py-4">
													<div className="text-sm text-gray-900">
														{formatDate(reclamo.fecha_registro)}
													</div>
													<div className={`text-xs ${
														reclamo.dias_restantes !== null && reclamo.dias_restantes < 3
															? 'text-red-600 font-medium'
															: 'text-gray-500'
													}`}>
														{reclamo.dias_restantes !== null ? (
															reclamo.dias_restantes >= 0 
																? `${reclamo.dias_restantes} días restantes`
																: `Vencido hace ${Math.abs(reclamo.dias_restantes)} días`
														) : 'Sin plazo'}
													</div>
												</td>
												<td className="px-6 py-4 text-right text-sm font-medium">
													<a
														href={`/admin/reclamos/${reclamo.id}`}
														className="inline-flex items-center gap-1 text-blue-600 hover:text-blue-900 font-medium transition"
													>
														Ver detalles
														<svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
														</svg>
													</a>
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
						)}
					</div>
				)}

				{/* Results Count */}
				{!loading && !error && filteredReclamos.length > 0 && (
					<div className="text-sm text-gray-600 text-center">
						Mostrando {filteredReclamos.length} de {reclamos.length} reclamos
					</div>
				)}
			</div>
		</AdminLayout>
	);
}