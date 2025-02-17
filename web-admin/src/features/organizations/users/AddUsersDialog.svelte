<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-management/UserRoleSelect.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    createAdminServiceAddOrganizationMemberUser,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import { page } from "$app/stores";

  export let open = false;
  export let email: string;
  export let role: string;
  export let isSuperUser: boolean;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const addOrganizationMemberUser =
    createAdminServiceAddOrganizationMemberUser();

  let failedInvites: string[] = [];

  async function handleCreate(
    newEmail: string,
    newRole: string,
    isSuperUser: boolean = false,
  ) {
    try {
      await $addOrganizationMemberUser.mutateAsync({
        organization: organization,
        data: {
          email: newEmail,
          role: newRole,
          superuserForceAccess: isSuperUser,
        },
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationInvitesQueryKey(organization),
      );

      email = "";
      role = "";
      isSuperUser = false;
    } catch (error) {
      console.error("Error adding user to organization", error);
      throw error;
    }
  }

  const formId = "add-user-form";

  const initialValues: {
    emails: string[];
    role: string;
  } = {
    emails: [""],
    role: "viewer",
  };
  const schema = yup(
    object({
      emails: array(
        string().matches(RFC5322EmailRegex, {
          excludeEmptyString: true,
          message: "Invalid email",
        }),
      ), // yup's email regex is too simple
      role: string().required(),
    }),
  );

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        failedInvites = [];
        let succeeded = [];
        let failed = [];

        if (!form.valid) return;
        const values = form.data;
        const emails = values.emails.map((e) => e.trim()).filter(Boolean);
        if (emails.length === 0) return;

        const results = await Promise.all(
          emails.map(async (email, index) => {
            try {
              await handleCreate(email, values.role, isSuperUser);
              return { index, email, success: true };
            } catch (error) {
              console.error("Error adding user to organization", error);
              return { index, email, success: false };
            }
          }),
        );

        results
          .sort((a, b) => a.index - b.index)
          .forEach(({ email, success }) => {
            if (success) {
              succeeded.push(email);
            } else {
              failed.push(email);
            }
          });

        // Only show success notification if any invites succeeded
        if (succeeded.length > 0) {
          eventBus.emit("notification", {
            type: "success",
            message: `Successfully invited ${succeeded.length} ${
              succeeded.length === 1 ? "person" : "people"
            } as ${values.role}`,
          });
        }

        // Show error notification if any invites failed
        if (failed.length > 0) {
          failedInvites = failed; // Store failed emails
          eventBus.emit("notification", {
            type: "error",
            message: `Failed to invite ${failed.length} ${
              failed.length === 1 ? "person" : "people"
            }`,
          });
        }

        // Close dialog after showing notifications
        if (failedInvites.length === 0) {
          open = false;
        }
      },
      validationMethod: "oninput",
    },
  );

  $: hasInvalidEmails = $form.emails.some(
    (e, i) => e.length > 0 && $errors.emails?.[i] !== undefined,
  );
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
    email = "";
    role = "";
    isSuperUser = false;
    failedInvites = [];
  }}
  onOpenChange={(open) => {
    if (!open) {
      email = "";
      role = "";
      isSuperUser = false;
      failedInvites = [];
    }
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Add users</DialogTitle>
    </DialogHeader>
    <form
      id={formId}
      on:submit|preventDefault={submit}
      class="w-full"
      use:enhance
    >
      <MultiInput
        id="emails"
        placeholder="Add emails, separated by commas"
        contentClassName="relative"
        bind:values={$form.emails}
        errors={$errors.emails}
        singular="email"
        plural="emails"
      >
        <div slot="within-input" class="flex items-center h-full">
          <UserRoleSelect bind:value={$form.role} />
        </div>
        <svelte:fragment slot="beside-input" let:hasSomeValue>
          <Button
            submitForm
            type="primary"
            form={formId}
            loading={$submitting}
            disabled={hasInvalidEmails || !hasSomeValue}
            forcedStyle="height: 32px !important;"
          >
            Invite
          </Button>
        </svelte:fragment>
      </MultiInput>
      {#if failedInvites.length > 0}
        <div class="text-sm text-red-500 py-2">
          Failed to invite {failedInvites.join(", ")}
        </div>
      {/if}
    </form>
  </DialogContent>
</Dialog>
