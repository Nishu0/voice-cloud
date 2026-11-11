<script lang="ts">
	// Ordered-dither sky: three flat colours, an 8×8 Bayer matrix, drifting cloud noise.
	// The centre of the frame is kept clear so whatever sits on top stays legible.
	// `fade` dissolves the sky into the page background at one edge.
	let { fade = 'none' }: { fade?: 'none' | 'bottom' | 'top' } = $props();

	let canvas: HTMLCanvasElement;

	const BAYER = [
		0, 32, 8, 40, 2, 34, 10, 42, 48, 16, 56, 24, 50, 18, 58, 26, 12, 44, 4, 36, 14, 46, 6, 38, 60,
		28, 52, 20, 62, 30, 54, 22, 3, 35, 11, 43, 1, 33, 9, 41, 51, 19, 59, 27, 49, 17, 57, 25, 15, 47,
		7, 39, 13, 45, 5, 37, 63, 31, 55, 23, 61, 29, 53, 21
	].map((v) => (v + 0.5) / 64);

	const abgr = (r: number, g: number, b: number) => (0xff000000 | (b << 16) | (g << 8) | r) >>> 0;
	const GROUND = abgr(11, 11, 18);
	const SKY_DEEP = abgr(40, 34, 176);
	const SKY = abgr(63, 56, 222);
	const CLOUD = abgr(176, 164, 248);
	const CLOUD_LIT = abgr(244, 241, 255);

	// Seeded value-noise lattice, so the sky looks the same on every load.
	const N = 256;
	const lattice = new Float32Array(N * N);
	let seed = 0x9e3779b9;
	for (let i = 0; i < lattice.length; i++) {
		seed = (seed + 0x6d2b79f5) | 0;
		let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
		t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
		lattice[i] = ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	}
	function noise(x: number, y: number) {
		const xi = Math.floor(x);
		const yi = Math.floor(y);
		let fx = x - xi;
		let fy = y - yi;
		fx = fx * fx * (3 - 2 * fx);
		fy = fy * fy * (3 - 2 * fy);
		const x0 = xi & 255;
		const x1 = (xi + 1) & 255;
		const y0 = (yi & 255) * N;
		const y1 = ((yi + 1) & 255) * N;
		const a = lattice[y0 + x0];
		const b = lattice[y0 + x1];
		const c = lattice[y1 + x0];
		const d = lattice[y1 + x1];
		return a + (b - a) * fx + (c - a + (a - b - c + d) * fx) * fy;
	}
	const fbm = (x: number, y: number) =>
		noise(x, y) * 0.5 +
		noise(x * 2.03, y * 2.03) * 0.27 +
		noise(x * 4.1, y * 4.1) * 0.15 +
		noise(x * 8.3, y * 8.3) * 0.08;
	const step = (lo: number, hi: number, v: number) => {
		const t = Math.min(1, Math.max(0, (v - lo) / (hi - lo)));
		return t * t * (3 - 2 * t);
	};

	$effect(() => {
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		const flip = fade === 'top';
		let cols = 0;
		let rows = 0;
		let cell = 3;
		let image: ImageData;
		let pixels: Uint32Array;

		function resize() {
			const { width, height } = canvas.getBoundingClientRect();
			if (!width || !height) return;
			cell = Math.max(3, Math.ceil(width / 560));
			cols = Math.ceil(width / cell);
			rows = Math.ceil(height / cell);
			canvas.width = cols;
			canvas.height = rows;
			image = ctx!.createImageData(cols, rows);
			pixels = new Uint32Array(image.data.buffer);
			draw(performance.now());
		}

		function draw(now: number) {
			if (!cols) return;
			const drift = now * 0.000012;
			for (let y = 0; y < rows; y++) {
				const g = flip ? 1 - y / rows : y / rows;
				const toGround = fade === 'none' ? 0 : step(0.6, 0.98, g);
				const depth = 0.15 + g * 0.8;
				const dy = (y / rows - 0.5) * 2;
				const ny = (y * cell) / 260;
				for (let x = 0; x < cols; x++) {
					const threshold = BAYER[(y & 7) * 8 + (x & 7)];
					let colour = depth > threshold ? SKY : SKY_DEEP;
					if (toGround > threshold) {
						colour = GROUND;
					} else {
						const dx = (x / cols - 0.5) * 2;
						const clear = step(0.62, 1.12, Math.sqrt(dx * dx * 0.85 + dy * dy * 1.5));
						if (clear > 0) {
							const n = fbm((x * cell) / 460 + drift, ny);
							if (step(0.47, 0.66, n) * clear * 1.25 > threshold) {
								colour = step(0.56, 0.8, n) * clear * 1.2 > threshold ? CLOUD_LIT : CLOUD;
							}
						}
					}
					pixels[y * cols + x] = colour;
				}
			}
			ctx!.putImageData(image, 0, 0);
		}

		const still = matchMedia('(prefers-reduced-motion: reduce)').matches;
		let onScreen = true;
		let frame = 0;
		let last = 0;
		function tick(now: number) {
			frame = requestAnimationFrame(tick);
			if (!onScreen || document.hidden || now - last < 80) return;
			last = now;
			draw(now);
		}

		const sizes = new ResizeObserver(resize);
		sizes.observe(canvas);
		const view = new IntersectionObserver(([entry]) => (onScreen = entry.isIntersecting));
		view.observe(canvas);
		if (!still) frame = requestAnimationFrame(tick);

		return () => {
			cancelAnimationFrame(frame);
			sizes.disconnect();
			view.disconnect();
		};
	});
</script>

<canvas bind:this={canvas} aria-hidden="true"></canvas>

<style>
	canvas {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		image-rendering: pixelated;
		background: linear-gradient(var(--sky-deep), var(--sky));
	}
</style>
