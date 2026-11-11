<script lang="ts">
	// Renders the product name as a dot-matrix: the text is rasterised small,
	// then each lit pixel becomes one square cell with a hairline gap.
	let { text }: { text: string } = $props();

	let wrap: HTMLElement;
	let canvas: HTMLCanvasElement;
	let ready = $state(false);

	$effect(() => {
		const label = text;
		const ctx = canvas.getContext('2d');
		const probe = document.createElement('canvas').getContext('2d', { willReadFrequently: true });
		if (!ctx || !probe) return;
		let cancelled = false;

		function draw() {
			const width = wrap.clientWidth;
			if (!width || cancelled) return;
			const fontPx = width < 520 ? 20 : 30;
			const font = `700 ${fontPx}px "JetBrains Mono Variable", ui-monospace, monospace`;
			probe!.font = font;
			const metrics = probe!.measureText(label);
			const ascent = Math.ceil(metrics.actualBoundingBoxAscent);
			const cols = Math.ceil(metrics.width) + 2;
			const rows = ascent + Math.ceil(metrics.actualBoundingBoxDescent) + 2;
			probe!.canvas.width = cols;
			probe!.canvas.height = rows;
			probe!.font = font;
			probe!.fillStyle = '#fff';
			probe!.fillText(label, 1, ascent + 1);
			const alpha = probe!.getImageData(0, 0, cols, rows).data;

			const dpr = window.devicePixelRatio || 1;
			const cell = Math.max(2, Math.floor((width * dpr) / cols));
			const dot = cell - Math.max(1, Math.round(cell * 0.16));
			canvas.width = cols * cell;
			canvas.height = rows * cell;
			canvas.style.width = `${(cols * cell) / dpr}px`;
			ctx!.clearRect(0, 0, canvas.width, canvas.height);
			for (let y = 0; y < rows; y++) {
				const shade = y / rows;
				ctx!.fillStyle = `rgb(${255 - shade * 38}, ${255 - shade * 46}, 255)`;
				for (let x = 0; x < cols; x++) {
					if (alpha[(y * cols + x) * 4 + 3] > 96) ctx!.fillRect(x * cell, y * cell, dot, dot);
				}
			}
			ready = true;
		}

		document.fonts.load('700 30px "JetBrains Mono Variable"').then(draw, draw);
		const sizes = new ResizeObserver(draw);
		sizes.observe(wrap);
		return () => {
			cancelled = true;
			sizes.disconnect();
		};
	});
</script>

<h1 bind:this={wrap} class:ready>
	<span>{text}</span>
	<canvas bind:this={canvas} aria-hidden="true"></canvas>
</h1>

<style>
	h1 {
		position: relative;
		display: grid;
		justify-items: center;
		width: min(88vw, 860px);
		margin: 0;
		font-size: clamp(2.5rem, 11vw, 7.5rem);
		font-weight: 700;
		line-height: 1;
		letter-spacing: -0.04em;
		white-space: nowrap;
		color: #fff;
	}
	span,
	canvas {
		grid-area: 1 / 1;
	}
	canvas {
		visibility: hidden;
		align-self: center;
	}
	/* Once the matrix is drawn the live text stays for assistive tech only. */
	.ready span {
		opacity: 0;
		user-select: none;
	}
	.ready canvas {
		visibility: visible;
	}
</style>
