<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import InfoTip from '#lib/components/InfoTip.svelte';
	import Terminal from 'phosphor-svelte/lib/TerminalWindow';
	import Eye from 'phosphor-svelte/lib/Eye';
	import EyeSlash from 'phosphor-svelte/lib/EyeSlash';

	let email = $state('');
	let password = $state('');
	let showPassword = $state(false);
	let message = $state('');

	function signIn(event: SubmitEvent) {
		event.preventDefault();
		password = '';
		message = 'Sign-in is not connected yet. Open the preview below.';
	}
</script>

<svelte:head><title>Sign in · Voice Cloud</title></svelte:head>

<div class="auth-page">
	<header class="auth-brand">
		<span class="brand-mark"><Terminal size={19} /></span><span
			>voice<span class="muted">cloud</span></span
		>
	</header>
	<section class="auth-panel" aria-labelledby="login-title">
		<div class="auth-heading">
			<span class="eyebrow">~/ auth</span>
			<div class="flex items-center gap-2">
				<h1 id="login-title">Sign in<span class="cursor">_</span></h1>
				<InfoTip
					label="About sign in"
					text="Login UI only. Credentials are not sent or stored; authentication will be connected later."
				/>
			</div>
		</div>
		<form method="POST" onsubmit={signIn} class="flex flex-col gap-4">
			<Field.Group>
				<Field.Field
					><Field.Label for="login-email">Email</Field.Label><Input
						id="login-email"
						name="email"
						type="email"
						autocomplete="username"
						placeholder="you@example.com"
						bind:value={email}
						required
					/></Field.Field
				>
				<Field.Field
					><Field.Label for="login-password">Password</Field.Label>
					<div class="password-control">
						<Input
							id="login-password"
							name="password"
							type={showPassword ? 'text' : 'password'}
							autocomplete="current-password"
							bind:value={password}
							required
						/><Button
							variant="ghost"
							size="icon-sm"
							aria-label={showPassword ? 'Hide password' : 'Show password'}
							aria-pressed={showPassword}
							onclick={() => (showPassword = !showPassword)}
							>{#if showPassword}<EyeSlash />{:else}<Eye />{/if}</Button
						>
					</div></Field.Field
				>
			</Field.Group>
			{#if message}<p class="auth-message" role="status">{message}</p>{/if}
			<Button type="submit" class="w-full">Sign in →</Button>
		</form>
		<div class="auth-preview">
			<span class="eyebrow">UI PREVIEW</span><Button href="/" variant="ghost" size="xs"
				>Open console ↗</Button
			>
		</div>
	</section>
	<footer class="auth-footer">self-hosted / v0.1</footer>
</div>
