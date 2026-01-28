// src/components/admin/ReclamoDetail.tsx
import { useEffect, useState } from 'react';
import AdminLayout from './AdminLayout';

const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:3000';

interface Reclamo {
	id: string;
	codigo_reclamo: string;
	tipo_solicitud: string;
	estado: string;
	nombre_completo: string;
	tipo_documento: string;
	numero_documento: string;
	telefono: string;
	email: string;
	domicilio: string | null;
	departamento: string | null;
	provincia: string | null;
	distrito: string | null;
	tipo_bien: string | null;
	monto_reclamado: number;
	descripcion_bien: string;
	detalle_reclamo: string;
	pedido_consumidor: string;
	fecha_registro: string;
	fecha_limite_respuesta: string;
	fecha_incidente: string;
	respuesta_empresa: string | null;
	respondido_por: string | null;
}

interface ReclamoDetailProps {
	id: string;
}

export default function ReclamoDetail({ id }: ReclamoDetailProps) {
	const [reclamo, setReclamo] = useState<Reclamo | null>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [nuevoEstado, setNuevoEstado] = useState('');
	const [comentario, setComentario] = useState('');
	const [respuesta, setRespuesta] = useState('');
	const [accionTomada, setAccionTomada] = useState('');
	const [compensacion, setCompensacion] = useState('');
	const [saving, setSaving] = useState(false);
	const [userRole, setUserRole] = useState('');
    const [mensajes, setMensajes] = useState<any[]>([]);
	const [nuevoMensaje, setNuevoMensaje] = useState('');
	const [sendingMessage, setSendingMessage] = useState(false);

	useEffect(() => {
		const userStr = localStorage.getItem('admin_user');
		if (userStr) {
			const user = JSON.parse(userStr);
			setUserRole(user.rol);
		}
		fetchReclamo();
	}, [id]);




useEffect(() => {
		if (id) fetchMensajes();
	}, [id]);

	const fetchMensajes = async () => {
		try {
			const token = localStorage.getItem('admin_token');
			const response = await fetch(`${API_URL}/api/admin/reclamos/${id}/mensajes`, {
				headers: { 'Authorization': `Bearer ${token}` },
			});
			const data = await response.json();
			if (data.success) {
				setMensajes(data.data);
			}
		} catch (err) {
			console.error('Error cargando mensajes:', err);
		}
	};

	const handleEnviarMensaje = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!nuevoMensaje.trim()) return;

		setSendingMessage(true);
		try {
			const token = localStorage.getItem('admin_token');
			const response = await fetch(`${API_URL}/api/admin/reclamos/${id}/mensaje`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${token}`,
				},
				body: JSON.stringify({ mensaje: nuevoMensaje }),
			});

			const data = await response.json();
			if (data.success) {
				setNuevoMensaje('');
				fetchMensajes();
			} else {
				alert(data.message || 'Error al enviar mensaje');
			}
		} catch (err: any) {
			alert(err.message || 'Error de conexión');
		} finally {
			setSendingMessage(false);
		}
	};


	const fetchReclamo = async () => {
		try {
			const token = localStorage.getItem('admin_token');
			if (!token) {
				window.location.href = '/admin/login';
				return;
			}

			// OJO: Agregamos "/admin" aquí
const response = await fetch(`${API_URL}/api/admin/reclamos/${id}`, {
    headers: { 'Authorization': `Bearer ${token}` },
});

			const data = await response.json();
			if (data.success) {
				setReclamo(data.data);
				setNuevoEstado(data.data.estado);
			} else {
				setError('Reclamo no encontrado');
			}
		} catch (err: any) {
			setError(err.message || 'Error de conexión');
		} finally {
			setLoading(false);
		}
	};

	const handleCambiarEstado = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!reclamo) return;

		setSaving(true);
		try {
			const token = localStorage.getItem('admin_token');
			const response = await fetch(`${API_URL}/api/admin/reclamos/${reclamo.id}/estado`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${token}`,
				},
				body: JSON.stringify({ estado: nuevoEstado, comentario }),
			});

			const data = await response.json();
			if (data.success) {
				alert('Estado actualizado correctamente');
				fetchReclamo();
				setComentario('');
			} else {
				alert(data.message || 'Error al actualizar estado');
			}
		} catch (err: any) {
			alert(err.message || 'Error de conexión');
		} finally {
			setSaving(false);
		}
	};

	const handleResponder = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!reclamo || respuesta.length < 10) {
			alert('La respuesta debe tener al menos 10 caracteres');
			return;
		}

		setSaving(true);
		try {
			const token = localStorage.getItem('admin_token');
			const response = await fetch(`${API_URL}/api/admin/reclamos/${reclamo.id}/respuesta`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${token}`,
				},
				body: JSON.stringify({
					respuesta_empresa: respuesta,
					accion_tomada: accionTomada || undefined,
					compensacion_ofrecida: compensacion || undefined,
				}),
			});

			const data = await response.json();
			if (data.success) {
				alert('Respuesta enviada correctamente');
				fetchReclamo();
				setRespuesta('');
				setAccionTomada('');
				setCompensacion('');
			} else {
				alert(data.message || 'Error al enviar respuesta');
			}
		} catch (err: any) {
			alert(err.message || 'Error de conexión');
		} finally {
			setSaving(false);
		}
	};

	const getEstadoBadgeClass = (estado: string) => {
		switch (estado) {
			case 'PENDIENTE': return 'bg-yellow-100 text-yellow-800';
			case 'EN_PROCESO': return 'bg-blue-100 text-blue-800';
			case 'RESUELTO': return 'bg-green-100 text-green-800';
			case 'CERRADO': return 'bg-gray-100 text-gray-800';
			default: return 'bg-gray-100 text-gray-800';
		}
	};

	if (loading) {
		return (
			<AdminLayout currentPage="reclamos">
				<div className="flex items-center justify-center py-12">
					<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
				</div>
			</AdminLayout>
		);
	}

	if (error || !reclamo) {
		return (
			<AdminLayout currentPage="reclamos">
				<div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
					{error || 'Reclamo no encontrado'}
				</div>
			</AdminLayout>
		);
	}

	return (
		<AdminLayout currentPage="reclamos">
			<div className="space-y-6">
				{/* Header */}
				<div className="flex items-center justify-between">
					<div>
						<a href="/admin/reclamos" className="text-sm text-blue-600 hover:text-blue-700 mb-2 inline-flex items-center gap-1">
							<svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
							</svg>
							Volver a la lista
						</a>
						<h1 className="text-3xl font-bold text-gray-900">{reclamo.codigo_reclamo}</h1>
						<p className="mt-1 text-gray-600">{reclamo.tipo_solicitud}</p>
					</div>
					<span className={`px-4 py-2 rounded-full text-sm font-semibold ${getEstadoBadgeClass(reclamo.estado)}`}>
						{reclamo.estado.replace('_', ' ')}
					</span>
				</div>

				<div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
					{/* Info Principal */}
					<div className="lg:col-span-2 space-y-6">
						{/* Datos del Cliente */}
						<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
							<h2 className="text-lg font-semibold text-gray-900 mb-4">Datos del Cliente</h2>
							<div className="grid grid-cols-2 gap-4">
								<div>
									<p className="text-sm text-gray-600">Nombre Completo</p>
									<p className="font-medium text-gray-900">{reclamo.nombre_completo}</p>
								</div>
								<div>
									<p className="text-sm text-gray-600">Documento</p>
									<p className="font-medium text-gray-900">{reclamo.tipo_documento}: {reclamo.numero_documento}</p>
								</div>
								<div>
									<p className="text-sm text-gray-600">Email</p>
									<p className="font-medium text-gray-900">{reclamo.email}</p>
								</div>
								<div>
									<p className="text-sm text-gray-600">Teléfono</p>
									<p className="font-medium text-gray-900">{reclamo.telefono}</p>
								</div>
								{reclamo.domicilio && (
									<div className="col-span-2">
										<p className="text-sm text-gray-600">Dirección</p>
										<p className="font-medium text-gray-900">{reclamo.domicilio}</p>
										{reclamo.distrito && <p className="text-sm text-gray-500">{reclamo.distrito}, {reclamo.provincia}, {reclamo.departamento}</p>}
									</div>
								)}
							</div>
						</div>

						{/* Detalle del Bien/Servicio */}
						{/* Detalle del Bien/Servicio */}
                    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 w-full">
                        <h2 className="text-lg font-semibold text-gray-900 mb-4">Bien/Servicio Reclamado</h2>
                        <div className="space-y-3">
                            <div>
                                <p className="text-sm text-gray-600">Tipo</p>
                                <p className="font-medium text-gray-900">{reclamo.tipo_bien || 'SERVICIO'}</p>
                            </div>
                            {reclamo.monto_reclamado > 0 && (
                                <div>
                                    <p className="text-sm text-gray-600">Monto Reclamado</p>
                                    <p className="font-medium text-gray-900">S/ {reclamo.monto_reclamado.toFixed(2)}</p>
                                </div>
                            )}
                            <div className="w-full">
                                <p className="text-sm text-gray-600">Descripción</p>
                                <p className="text-gray-900 whitespace-pre-wrap break-words">{reclamo.descripcion_bien}</p>
                            </div>
                        </div>
                    </div>

                    {/* Detalle del Reclamo */}
                    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 w-full">
                        <h2 className="text-lg font-semibold text-gray-900 mb-4">Detalle del {reclamo.tipo_solicitud}</h2>
                        <div className="space-y-3">
                            <div>
                                <p className="text-sm text-gray-600">Fecha del Incidente</p>
                                <p className="font-medium text-gray-900">{new Date(reclamo.fecha_incidente).toLocaleDateString('es-PE')}</p>
                            </div>
                            <div className="w-full">
                                <p className="text-sm text-gray-600">Descripción del Problema</p>
                                <p className="text-gray-900 whitespace-pre-wrap break-words">{reclamo.detalle_reclamo}</p>
                            </div>
                            <div className="w-full">
                                <p className="text-sm text-gray-600">Pedido del Consumidor</p>
                                <p className="text-gray-900 whitespace-pre-wrap break-words">{reclamo.pedido_consumidor}</p>
                            </div>
                        </div>
                    </div>

                    {/* Respuesta (si existe) */}
                    {reclamo.respuesta_empresa && (
                        <div className="bg-green-50 rounded-lg border border-green-200 p-6 w-full">
                            <h2 className="text-lg font-semibold text-green-900 mb-4">✅ Respuesta de la Empresa</h2>
                            <div className="space-y-3">
                                <div className="w-full">
                                    <p className="text-sm text-green-700">Respuesta</p>
                                    <p className="text-gray-900 whitespace-pre-wrap break-words">{reclamo.respuesta_empresa}</p>
                                </div>
                                {reclamo.respondido_por && (
                                    <p className="text-xs text-green-600 break-words">Respondido por: {reclamo.respondido_por}</p>
                                )}
                            </div>
                        </div>
                    )}
                </div>

					{/* Sidebar - Acciones */}
					<div className="space-y-6">
						{/* Cambiar Estado */}
						<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
							<h2 className="text-lg font-semibold text-gray-900 mb-4">Cambiar Estado</h2>
							<form onSubmit={handleCambiarEstado} className="space-y-4">
								<div>
									<label className="block text-sm font-medium text-gray-700 mb-2">Estado</label>
									<select
										value={nuevoEstado}
										onChange={(e) => setNuevoEstado(e.target.value)}
										className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
										required
									>
										<option value="PENDIENTE">Pendiente</option>
										<option value="EN_PROCESO">En Proceso</option>
										<option value="RESUELTO">Resuelto</option>
										{userRole === 'ADMIN' && <option value="CERRADO">Cerrado</option>}
									</select>
								</div>
								<div>
									<label className="block text-sm font-medium text-gray-700 mb-2">Comentario (opcional)</label>
									<textarea
										value={comentario}
										onChange={(e) => setComentario(e.target.value)}
										rows={3}
										className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
										placeholder="Comentario interno..."
									/>
								</div>
								<button
									type="submit"
									disabled={saving || nuevoEstado === reclamo.estado}
									className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white font-semibold py-2 px-4 rounded-lg transition"
								>
									{saving ? 'Guardando...' : 'Actualizar Estado'}
								</button>
							</form>
						</div>

						{/* Responder al Cliente */}
						{!reclamo.respuesta_empresa && (
							<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
								<h2 className="text-lg font-semibold text-gray-900 mb-4">Responder al Cliente</h2>
								<form onSubmit={handleResponder} className="space-y-4">
									<div>
										<label className="block text-sm font-medium text-gray-700 mb-2">Respuesta *</label>
										<textarea
											value={respuesta}
											onChange={(e) => setRespuesta(e.target.value)}
											rows={4}
											className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
											placeholder="Detalle de la respuesta al cliente..."
											required
											minLength={10}
										/>
										<p className="text-xs text-gray-500 mt-1">Mínimo 10 caracteres</p>
									</div>
									<div>
										<label className="block text-sm font-medium text-gray-700 mb-2">Acción Tomada</label>
										<input
											type="text"
											value={accionTomada}
											onChange={(e) => setAccionTomada(e.target.value)}
											className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
											placeholder="Ej: Reemplazo del producto"
										/>
									</div>
									<div>
										<label className="block text-sm font-medium text-gray-700 mb-2">Compensación</label>
										<input
											type="text"
											value={compensacion}
											onChange={(e) => setCompensacion(e.target.value)}
											className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
											placeholder="Ej: Cupón de descuento"
										/>
									</div>
									<button
										type="submit"
										disabled={saving || respuesta.length < 10}
										className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-300 text-white font-semibold py-2 px-4 rounded-lg transition"
									>
										{saving ? 'Enviando...' : 'Enviar Respuesta'}
									</button>
								</form>
							</div>
						)}

						{/* Info de Plazos */}
						<div className="bg-blue-50 rounded-lg border border-blue-200 p-4">
							<h3 className="text-sm font-semibold text-blue-900 mb-2">📅 Plazos</h3>
							<div className="space-y-2 text-sm">
								<div className="flex justify-between">
									<span className="text-blue-700">Registrado:</span>
									<span className="font-medium text-blue-900">{new Date(reclamo.fecha_registro).toLocaleDateString('es-PE')}</span>
								</div>
								<div className="flex justify-between">
									<span className="text-blue-700">Límite respuesta:</span>
									<span className="font-medium text-blue-900">{new Date(reclamo.fecha_limite_respuesta).toLocaleDateString('es-PE')}</span>
								</div>
							</div>
						</div>
					</div>

					{/* Sección de Mensajes */}
					<div className="space-y-6">
						<div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
							<h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
								<svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
								</svg>
								Conversación con Cliente
							</h2>

							{/* Lista de Mensajes */}
							<div className="space-y-3 mb-4 max-h-96 overflow-y-auto">
								{mensajes.length === 0 ? (
									<p className="text-gray-500 text-center py-4 text-sm">No hay mensajes aún</p>
								) : (
									mensajes.map((msg) => (
										<div
											key={msg.id}
											className={`p-3 rounded-lg ${
												msg.tipo_mensaje === 'CLIENTE'
													? 'bg-purple-50 border-l-4 border-purple-400'
													: 'bg-blue-50 border-l-4 border-blue-400'
											}`}
										>
											<div className="flex justify-between items-start mb-1">
												<span className={`text-xs font-semibold ${
													msg.tipo_mensaje === 'CLIENTE' ? 'text-purple-700' : 'text-blue-700'
												}`}>
													{msg.tipo_mensaje === 'CLIENTE' ? '👤 Cliente' : '🏢 CODEPLEX'}
												</span>
												<span className="text-xs text-gray-500">
													{new Date(msg.fecha_mensaje).toLocaleDateString('es-PE', {
														day: '2-digit',
														month: '2-digit',
														year: 'numeric',
														hour: '2-digit',
														minute: '2-digit'
													})}
												</span>
											</div>
											<p className="text-sm text-gray-800 whitespace-pre-wrap">{msg.mensaje}</p>
										</div>
									))
								)}
							</div>

							{/* Formulario de Mensaje */}
							<form onSubmit={handleEnviarMensaje} className="space-y-3">
								<textarea
									value={nuevoMensaje}
									onChange={(e) => setNuevoMensaje(e.target.value)}
									rows={3}
									maxLength={1000}
									placeholder="Escribe un mensaje al cliente..."
									className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 text-sm resize-none"
								/>
								<div className="flex items-center justify-between">
									<span className="text-xs text-gray-500">{nuevoMensaje.length}/1000</span>
									<button
										type="submit"
										disabled={sendingMessage || !nuevoMensaje.trim()}
										className="px-4 py-2 bg-purple-600 hover:bg-purple-700 disabled:bg-gray-300 text-white text-sm font-semibold rounded-lg transition"
									>
										{sendingMessage ? 'Enviando...' : 'Enviar Mensaje'}
									</button>
								</div>
							</form>
						</div>
					</div>
				</div>
			</div>
		</AdminLayout>
	);
}