#!/bin/bash

# Crear FirmaDigital.astro
cat > frontend/src/components/FirmaDigital.astro << 'EOFCOMP'
---
// Componente de Firma Digital
---
<div class="firma-digital-container">
  <label class="block text-sm font-medium text-gray-700 mb-2">
    Firma Digital <span class="text-red-500">*</span>
  </label>
  <div class="bg-white border-2 border-gray-300 rounded-lg p-4">
    <p class="text-sm text-gray-600 mb-3">
      Firme con su dedo o mouse en el recuadro para dar conformidad:
    </p>
    <canvas
      id="signature-canvas"
      class="border-2 border-gray-400 rounded cursor-crosshair bg-white w-full"
      width="600"
      height="200"
    ></canvas>
    <div class="flex gap-3 mt-3">
      <button
        type="button"
        id="clear-signature"
        class="px-4 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg transition"
      >
        🗑️ Limpiar
      </button>
      <button
        type="button"
        id="undo-signature"
        class="px-4 py-2 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg transition"
      >
        ↶ Deshacer
      </button>
    </div>
    <p class="text-xs text-gray-500 mt-2">
      ℹ️ Su firma digital tiene validez legal según Ley N° 27269
    </p>
  </div>
  <input type="hidden" id="firma_digital" name="firma_digital" required />
</div>

<script>
  import SignaturePad from 'https://cdn.jsdelivr.net/npm/signature_pad@5.0.0/+esm';
  
  const canvas = document.getElementById('signature-canvas') as HTMLCanvasElement;
  const hiddenInput = document.getElementById('firma_digital') as HTMLInputElement;
  
  if (canvas && hiddenInput) {
    const signaturePad = new SignaturePad(canvas, {
      backgroundColor: 'rgb(255, 255, 255)',
      penColor: 'rgb(0, 0, 0)'
    });
    
    signaturePad.addEventListener('endStroke', () => {
      if (!signaturePad.isEmpty()) {
        hiddenInput.value = signaturePad.toDataURL('image/png');
      }
    });
    
    document.getElementById('clear-signature')?.addEventListener('click', () => {
      signaturePad.clear();
      hiddenInput.value = '';
    });
    
    document.getElementById('undo-signature')?.addEventListener('click', () => {
      const data = signaturePad.toData();
      if (data && data.length > 0) {
        data.pop();
        signaturePad.fromData(data);
        hiddenInput.value = signaturePad.isEmpty() ? '' : signaturePad.toDataURL('image/png');
      }
    });
  }
</script>
EOFCOMP

echo "✅ FirmaDigital.astro creado"
