<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import InfoTip from '#lib/components/InfoTip.svelte';
	import Terminal from 'phosphor-svelte/lib/TerminalWindow';
	import Eye from 'phosphor-svelte/lib/Eye';
	import EyeSlash from 'phosphor-svelte/lib/EyeSlash';

	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	let { form }: PageProps = $props();
	let pending = $state(false);
	let email = $state('');
	let password = $state('');
	let showPassword = $state(false);
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
				<InfoTip label="About sign in" text="Use the email and password created during setup." />
			</div>
		</div>
		<form
			method="POST"
			use:enhance={() => {
				pending = true;
				return async ({ update }) => {
					password = '';
					await update();
					pending = false;
				};
			}}
			class="flex flex-col gap-4"
		>
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
			{#if form?.error}<p class="auth-message" role="alert">{form.error}</p>{/if}
			<Button type="submit" class="w-full" disabled={pending}
				>{pending ? 'Signing in...' : 'Sign in →'}</Button
			>
		</form>
	</section>
	<footer class="auth-footer">self-hosted / v0.1</footer>
</div>
