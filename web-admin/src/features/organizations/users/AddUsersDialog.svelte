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
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-invite/UserRoleSelect.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation";
  import { eventBus } from "@rilldata/events";
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
      open = false;

      eventBus.emit("notification", { message: "User added to organization" });
    } catch (error) {
      console.error("Error adding user to organization", error);
      eventBus.emit("notification", {
        message: "Error adding user to organization",
        type: "error",
      });
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
        if (!form.valid) return;
        const values = form.data;
        const emails = values.emails.map((e) => e.trim()).filter(Boolean);
        if (emails.length === 0) return;

        const succeeded = [];
        let errored = false;
        await Promise.all(
          emails.map(async (email) => {
            try {
              await handleCreate(email, values.role, isSuperUser);
              succeeded.push(email);
            } catch (error) {
              console.error("Error adding user to organization", error);
              errored = true;
            }
          }),
        );

        eventBus.emit("notification", {
          type: "success",
          message: `Invited ${succeeded.length} ${succeeded.length === 1 ? "person" : "people"} as ${values.role}`,
        });

        if (errored) {
          eventBus.emit("notification", {
            type: "error",
            message:
              "Some invitations could not be sent. Please check the email addresses and try again.",
          });
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
  }}
  onOpenChange={(open) => {
    if (!open) {
      email = "";
      role = "";
      isSuperUser = false;
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
    </form>
  </DialogContent>
</Dialog>
