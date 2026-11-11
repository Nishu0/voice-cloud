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

<svelte:head><title>Create account · Voice Cloud</title></svelte:head>

<div class="auth-page">
	<header class="auth-brand">
		<span class="brand-mark"><Terminal size={19} /></span><span
			>voice<span class="muted">cloud</span></span
		>
	</header>
	<section class="auth-panel" aria-labelledby="login-title">
		<div class="auth-heading">
			<span class="eyebrow">~/ setup</span>
			<div class="flex items-center gap-2">
				<h1 id="login-title">Create account<span class="cursor">_</span></h1>
				<InfoTip
					label="About sign in"
					text="Creates the first administrator. Setup closes after this account is created."
				/>
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
			<Field.Group
				><Field.Field
					><Field.Label for="signup-name">Name</Field.Label><Input
						id="signup-name"
						name="name"
						autocomplete="name"
						required
						maxlength={100}
					/></Field.Field
				>
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
					><div class="flex items-center gap-1">
						<Field.Label for="login-password">Password</Field.Label><InfoTip
							text="Use at least 12 characters."
							label="Password requirements"
						/>
					</div>
					<div class="password-control">
						<Input
							id="login-password"
							name="password"
							type={showPassword ? 'text' : 'password'}
							autocomplete="new-password"
							bind:value={password}
							minlength={12}
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
				>{pending ? 'Creating account...' : 'Create account →'}</Button
			>
		</form>
	</section>
	<footer class="auth-footer">self-hosted / v0.1</footer>
</div>
