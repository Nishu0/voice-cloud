<script lang="ts">
	import type { PageProps } from './$types';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import ConfigurationTable from '#lib/components/ConfigurationTable.svelte';
	import InfoTip from '#lib/components/InfoTip.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { enhance } from '$app/forms';
	let { data, form }: PageProps = $props();
	const title = $derived(
		(
			{
				connections: 'Providers',
				deployments: 'Deployments',
				policies: 'Routing policies',
				models: 'Models',
			} as Record<string, string>
		)[data.resource],
	);
	const records = $derived(
		data.resource === 'connections'
			? data.connections
			: data.resource === 'deployments'
				? data.deployments
				: data.resource === 'policies'
					? data.policies
					: data.models,
	);
</script>

<PageHeader {title} path={`manage / ${data.resource}`} />
<ConfigurationTable resource={data.resource} configuration={data} {records} />
{#if data.resource === 'deployments'}<section class="section">
		<div class="section-heading">
			<h2>Health & languages</h2>
			<InfoTip
				text="Reachability is checked by the worker. It does not measure inference availability."
			/>
		</div>
		{#each data.deployments as deployment}<div class="panel mb-3">
				<div class="panel-heading">
					<h2>{deployment.name}</h2>
					<span class={data.health[deployment.id]?.status === 'healthy' ? 'positive' : 'muted'}
						>{data.health[deployment.id]?.status ?? 'unknown'}
						<span class="muted">/ {data.health[deployment.id]?.error_code ?? 'checked'}</span></span
					>
				</div>
				<div class="panel-body">
					<div class="flex flex-wrap gap-2 mb-2">
						{#each data.languages[deployment.id] ?? [] as language}<form
								method="POST"
								action="?/language"
								use:enhance
							>
								<input type="hidden" name="id" value={deployment.id} /><input
									type="hidden"
									name="language"
									value={language.language_tag}
								/><input type="hidden" name="remove" value="true" /><Button
									variant="outline"
									size="xs"
									type="submit"
									disabled={data.user?.role !== 'admin'}>{language.language_tag} ×</Button
								>
							</form>{/each}
					</div>
					{#if data.user?.role === 'admin'}<form
							method="POST"
							action="?/language"
							use:enhance
							class="flex gap-2"
						>
							<input type="hidden" name="id" value={deployment.id} /><Input
								name="language"
								aria-label={`Language for ${deployment.name}`}
								placeholder="en-US"
								required
								class="max-w-36"
							/><Button variant="outline" size="sm" type="submit">Add language</Button>
						</form>{/if}
				</div>
			</div>{/each}
	</section>{/if}
{#if data.resource === 'policies'}<section class="section">
		<h2 class="mb-3">Routing targets</h2>
		{#each data.policies as policy}<div class="panel mb-3">
				<div class="panel-heading">
					<h2>{policy.name}</h2>
					<span class="eyebrow">{policy.task.toUpperCase()}</span>
				</div>
				<div class="panel-body">
					{#each data.targets[policy.id] ?? [] as target}<form
							method="POST"
							action="?/target"
							use:enhance
							class="flex items-center gap-2 mb-2"
						>
							<input type="hidden" name="id" value={policy.id} /><input
								type="hidden"
								name="deployment_id"
								value={target.deployment_id}
							/><input type="hidden" name="remove" value="true" /><span class="muted"
								>#{target.priority}</span
							><span>{data.deployments.find((d) => d.id === target.deployment_id)?.name}</span
							>{#if data.user?.role === 'admin'}<Button variant="ghost" size="xs" type="submit"
									>Remove ×</Button
								>{/if}
						</form>{/each}{#if data.user?.role === 'admin'}<form
							method="POST"
							action="?/target"
							use:enhance
							class="flex flex-wrap gap-2"
						>
							<input type="hidden" name="id" value={policy.id} /><select
								name="deployment_id"
								aria-label={`Target for ${policy.name}`}
								class="native-select max-w-64"
								required
								><option value="">Select deployment...</option
								>{#each data.deployments.filter((d) => data.models.find((m) => m.id === d.model_id)?.task === policy.task) as deployment}<option
										value={deployment.id}>{deployment.name}</option
									>{/each}</select
							><Input
								name="priority"
								type="number"
								min="0"
								value={0}
								aria-label={`Priority for ${policy.name}`}
								class="max-w-20"
							/><Button variant="outline" size="sm" type="submit">Set target</Button>
						</form>{/if}
				</div>
			</div>{/each}
	</section>{/if}
{#if form?.error}<p class="negative mt-3" role="alert">{form.error}</p>{/if}
