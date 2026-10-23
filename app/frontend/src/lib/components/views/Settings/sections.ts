import type { Icons } from '$lib/system/Icon/types';
import { isLocalEdition } from '$lib/stores/envStore';

export type SettingsSectionId =
	| 'workspace'
	| 'users'
	| 'roles'
	| 'groups'
	| 'api_keys'
	| 'connections'
	| 'theme'
	| 'config';

export type SettingsSection = {
	id: SettingsSectionId;
	label: string;
	icon: Icons;
	/**
	 * Permission required to view the section. Undefined = always accessible
	 * (personal sections like theme/config apply across all workspaces).
	 */
	action?: string;
	/** The local edition leaves it out. */
	needsServer?: true;
};

/** Single source of truth for Settings sections (nav, tab labels, resource menu). */
export const settingsSections: SettingsSection[] = [
	{ id: 'workspace', label: 'Workspace', icon: 'folder', action: 'workspace/settings.write' },
	{
		id: 'users',
		label: 'Users',
		icon: 'users',
		action: 'workspace/users.manage',
		needsServer: true
	},
	{ id: 'roles', label: 'Roles', icon: 'roles', action: 'workspace/roles.manage' },
	{ id: 'groups', label: 'Groups', icon: 'users', action: 'workspace/groups.manage' },
	{
		id: 'api_keys',
		label: 'API keys',
		icon: 'key',
		action: 'workspace/api-keys.manage',
		needsServer: true
	},
	// No action: a connection is administrated per connection rather than
	// workspace-wide, so someone who manages one but is not a workspace admin
	// still belongs here. Each row decides for itself whether it offers Revoke.
	{ id: 'connections', label: 'Connections', icon: 'db', needsServer: true },
	{ id: 'theme', label: 'Theme', icon: 'theme' },
	{ id: 'config', label: 'Config', icon: 'cog' }
];

export const offeredSettingsSections = settingsSections.filter(
	(s) => !(isLocalEdition && s.needsServer)
);

export const settingsSectionLabels: Record<string, string> = Object.fromEntries(
	settingsSections.map((s) => [s.id, s.label])
);
