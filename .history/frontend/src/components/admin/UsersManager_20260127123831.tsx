import { useState, useEffect } from 'react';

const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:3000';

interface User {
  id: string;
  email: string;
  nombre_completo: string;
  rol: 'ADMIN' | 'SOPORTE';
  activo: boolean;
  fecha_creacion: string;
}

export default function UsersManager() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  // Modals
  const [showCreate, setShowCreate] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [passUser, setPassUser] = useState<User | null>(null);

  // Forms
  const [formData, setFormData] = useState({ email: '', nombre_completo: '', password: '', rol: 'SOPORTE' });
  const [newPass, setNewPass] = useState('');

  const token = localStorage.getItem('admin_token');

  const fetchUsers = async () => {
    try {
      const res = await fetch(`${API_URL}/api/admin/usuarios`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await res.json();
      if (data.success) setUsers(data.data || []);
    } catch (err) { setError('Error cargando usuarios'); } 
    finally { setLoading(false); }
  };

  useEffect(() => { fetchUsers(); }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if(!confirm('¿Crear usuario?')) return;
    try {
      const res = await fetch(`${API_URL}/api/admin/usuarios`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify(formData)
      });
      if (res.ok) {
        setShowCreate(false);
        setFormData({ email: '', nombre_completo: '', password: '', rol: 'SOPORTE' });
        fetchUsers();
      } else { alert('Error al crear'); }
    } catch (err) { alert('Error de conexión'); }
  };

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingUser) return;
    try {
      const res = await fetch(`${API_URL}/api/admin/usuarios/${editingUser.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({
          nombre_completo: editingUser.nombre_completo,
          rol: editingUser.rol,
          activo: editingUser.activo
        })
      });
      if (res.ok) { setEditingUser(null); fetchUsers(); }
    } catch (err) { alert('Error al editar'); }
  };

  const handleChangePass = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!passUser) return;
    try {
      const res = await fetch(`${API_URL}/api/admin/usuarios/${passUser.id}/password`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ new_password: newPass })
      });
      if (res.ok) { 
        alert('Contraseña actualizada'); 
        setPassUser(null); 
        setNewPass(''); 
      }
    } catch (err) { alert('Error al cambiar contraseña'); }
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800">Gestión de Usuarios</h1>
        <button onClick={() => setShowCreate(true)} className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
          + Nuevo Usuario
        </button>
      </div>

      {loading ? <p>Cargando...</p> : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Usuario</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Rol</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Estado</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Acciones</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {users.map(u => (
                <tr key={u.id}>
                  <td className="px-6 py-4">
                    <div className="text-sm font-medium text-gray-900">{u.nombre_completo}</div>
                    <div className="text-sm text-gray-500">{u.email}</div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${u.rol === 'ADMIN' ? 'bg-purple-100 text-purple-800' : 'bg-green-100 text-green-800'}`}>
                      {u.rol}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${u.activo ? 'bg-blue-100 text-blue-800' : 'bg-red-100 text-red-800'}`}>
                      {u.activo ? 'Activo' : 'Inactivo'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right text-sm font-medium space-x-2">
                    <button onClick={() => setEditingUser(u)} className="text-indigo-600 hover:text-indigo-900">Editar</button>
                    <button onClick={() => setPassUser(u)} className="text-yellow-600 hover:text-yellow-900">Contraseña</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* MODAL CREAR */}
      {showCreate && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">Nuevo Usuario</h2>
            <form onSubmit={handleCreate} className="space-y-4">
              <input required type="email" placeholder="Email" className="w-full border p-2 rounded"
                value={formData.email} onChange={e => setFormData({...formData, email: e.target.value})} />
              <input required type="text" placeholder="Nombre Completo" className="w-full border p-2 rounded"
                value={formData.nombre_completo} onChange={e => setFormData({...formData, nombre_completo: e.target.value})} />
              <input required type="password" placeholder="Contraseña" className="w-full border p-2 rounded"
                value={formData.password} onChange={e => setFormData({...formData, password: e.target.value})} />
              <select className="w-full border p-2 rounded" value={formData.rol} onChange={e => setFormData({...formData, rol: e.target.value as any})}>
                <option value="SOPORTE">Soporte</option>
                <option value="ADMIN">Administrador</option>
              </select>
              <div className="flex justify-end gap-2 mt-4">
                <button type="button" onClick={() => setShowCreate(false)} className="text-gray-500">Cancelar</button>
                <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded">Guardar</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL EDITAR */}
      {editingUser && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">Editar Usuario</h2>
            <form onSubmit={handleEdit} className="space-y-4">
              <div>
                <label className="block text-sm text-gray-600">Nombre</label>
                <input required type="text" className="w-full border p-2 rounded"
                  value={editingUser.nombre_completo} onChange={e => setEditingUser({...editingUser, nombre_completo: e.target.value})} />
              </div>
              <div>
                <label className="block text-sm text-gray-600">Rol</label>
                <select className="w-full border p-2 rounded" value={editingUser.rol} onChange={e => setEditingUser({...editingUser, rol: e.target.value as any})}>
                  <option value="SOPORTE">Soporte</option>
                  <option value="ADMIN">Administrador</option>
                </select>
              </div>
              <div className="flex items-center gap-2">
                <input type="checkbox" id="activeCheck" checked={editingUser.activo} onChange={e => setEditingUser({...editingUser, activo: e.target.checked})} />
                <label htmlFor="activeCheck">Usuario Activo</label>
              </div>
              <div className="flex justify-end gap-2 mt-4">
                <button type="button" onClick={() => setEditingUser(null)} className="text-gray-500">Cancelar</button>
                <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded">Actualizar</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* MODAL PASSWORD */}
      {passUser && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">Cambiar Contraseña: {passUser.nombre_completo}</h2>
            <form onSubmit={handleChangePass} className="space-y-4">
              <input required type="password" placeholder="Nueva Contraseña" className="w-full border p-2 rounded"
                value={newPass} onChange={e => setNewPass(e.target.value)} />
              <div className="flex justify-end gap-2 mt-4">
                <button type="button" onClick={() => setPassUser(null)} className="text-gray-500">Cancelar</button>
                <button type="submit" className="bg-yellow-600 text-white px-4 py-2 rounded">Cambiar</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}