<script lang="ts">
  import { createAdminServiceAddProjectMemberUser } from "@rilldata/web-admin/client";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-invite/UserRoleSelect.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation";
  import { eventBus } from "@rilldata/events";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";

  export let organization: string;
  export let project: string;
  export let onInvite: () => void = () => {};

  const userInvite = createAdminServiceAddProjectMemberUser();

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
              await $userInvite.mutateAsync({
                organization,
                project,
                data: {
                  email,
                  role: values.role,
                },
              });
              succeeded.push(email);
            } catch {
              errored = true;
            }
          }),
        );

        eventBus.emit("notification", {
          type: "success",
          message: `Invited ${succeeded.length} ${succeeded.length === 1 ? "person" : "people"} as ${values.role}`,
        });
        onInvite();
        if (errored) {
          // TODO: there no mocks for this yet, but will be added in future.
          //       the challenge here is how to show it for all the emails that fail
        }
      },
      validationMethod: "oninput",
    },
  );

  $: hasInvalidEmails = $form.emails.some(
    (e, i) => e.length > 0 && $errors.emails?.[i] !== undefined,
  );
</script>

<form
  id="user-invite-form"
  on:submit|preventDefault={submit}
  class="w-full"
  use:enhance
>
  <MultiInput
    id="emails"
    placeholder="Invite by email, separated by commas"
    contentClassName="relative"
    bind:values={$form.emails}
    errors={$errors.emails}
    singular="email"
    plural="emails"
  >
    <div slot="within-input" class="h-full items-center flex">
      <UserRoleSelect bind:value={$form.role} />
    </div>
    <svelte:fragment slot="beside-input" let:hasSomeValue>
      <Button
        submitForm
        type="primary"
        form="user-invite-form"
        loading={$submitting}
        disabled={hasInvalidEmails || !hasSomeValue}
        forcedStyle="height: 32px !important;"
      >
        Invite
      </Button>
    </svelte:fragment>
  </MultiInput>
</form>
