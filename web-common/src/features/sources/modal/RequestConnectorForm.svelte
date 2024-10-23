<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { eventBus } from "@rilldata/events";
  import { createEventDispatcher } from "svelte";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";

  const dispatch = createEventDispatcher();

  const FORM_ID = "1P9sP1jxjFcMqDzxsweIrZiU7pFUgRY452S3Nk7cEeao";
  const GOOGLE_FORM_ENDPOINT = `https://docs.google.com/forms/d/${FORM_ID}`;
  const REQUEST_FIELD_ID = "entry.849552298";
  const EMAIL_FIELD_ID = "entry.516049603";

  const initialValues = {
    request: "",
    email: "",
  };

  const validationSchema = object({
    request: string().required("Required"),
    email: string().email("Invalid email"),
  });

  const { form, enhance, submit, errors, submitting } = superForm(
    defaults(initialValues, yup(validationSchema)),
    {
      SPA: true,
      validators: yup(validationSchema),
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;

        // Following the approach here: https://stackoverflow.com/questions/51995070/post-data-to-a-google-form-with-ajax
        const submitFormEndpoint = `${GOOGLE_FORM_ENDPOINT}/formResponse?${REQUEST_FIELD_ID}=${values.request}&${EMAIL_FIELD_ID}=${values.email}&submit=Submit`;

        try {
          await fetch(submitFormEndpoint, {
            method: "GET",
            mode: "no-cors",
            headers: {
              "Content-Type": "application/x-www-form-urlencoded",
            },
          });
          dispatch("close");
          eventBus.emit("notification", {
            message: "Thanks for your request!",
          });
        } catch (e) {
          console.error(e);
        }
      },
    },
  );
</script>

<form on:submit|preventDefault={submit} id="request-connector-form" use:enhance>
  <span class="text-slate-500 text-sm">
    Don't see the connector you're looking for? Let us know what we're missing!
  </span>
  <Input
    id="request"
    label="Connector"
    placeholder="Your data source"
    errors={$errors.request}
    bind:value={$form.request}
    alwaysShowError
  />
  <Input
    id="email"
    label="Optionally, we can let you know when the connector is available."
    placeholder="Your email address"
    errors={$errors.email}
    bind:value={$form.email}
  />
  <div class="flex gap-x-2">
    <div class="grow" />
    <Button on:click={() => dispatch("back")} type="secondary">Back</Button>
    <Button
      type="primary"
      submitForm
      form="request-connector-form"
      disabled={$submitting}
    >
      Request connector
    </Button>
  </div>
</form>

<style lang="postcss">
  form {
    @apply flex flex-col gap-y-4;
  }
</style>
