<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import {
    useDirectoryNamesInDirectory,
    useFileNamesInDirectory,
  } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { defaults, setError, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { object, string } from "yup";
  import { runtime } from "../../runtime-client/runtime-store";
  import { renameFileArtifact } from "./actions";
  import { removeLeadingSlash } from "./entity-mappers";
  import {
    INVALID_NAME_MESSAGE,
    VALID_NAME_PATTERN,
    isDuplicateName,
    splitFolderAndFileName,
  } from "@rilldata/utils";

  export let closeModal: () => void;
  export let filePath: string;
  export let isDir: boolean;

  let error: string;

  const [folderName, fileName] = splitFolderAndFileName(filePath);

  const validationSchema = object({
    newName: string()
      .required("Enter a name!")
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE),
  });

  const initialValues = {
    newName: fileName,
  };

  const {
    form: superform,
    enhance,
    submit,
    errors,
  } = superForm(defaults(initialValues, yup(validationSchema)), {
    SPA: true,
    validators: yup(validationSchema),
    validationMethod: "onsubmit",
    async onUpdate({ form }) {
      if (!form.valid) return;

      const values = form.data;

      if (values.newName === fileName) {
        closeModal();
        return;
      }

      if (
        isDir &&
        isDuplicateName(
          values?.newName,
          fileName,
          $existingDirectories?.data ?? [],
        )
      ) {
        error = `An existing folder with name ${values.newName} already exists`;
        return setError(form, "newName", error);
      }

      if (
        !isDir &&
        isDuplicateName(
          values?.newName,
          fileName,
          $fileNamesInDirectory?.data ?? [],
        )
      ) {
        error = `Name ${values.newName} is already in use`;

        return setError(form, "newName", error);
      }
      try {
        const newPath = (folderName ? `${folderName}/` : "") + values.newName;
        await renameFileArtifact(runtimeInstanceId, filePath, newPath);
        if (isDir) {
          if (
            $page.url.pathname.startsWith(
              `/files/${removeLeadingSlash(filePath)}`,
            )
          ) {
            // if the file focused has the dir then replace the dir path to the new one
            void goto(
              $page.url.pathname.replace(
                `/files/${removeLeadingSlash(filePath)}`,
                `/files/${removeLeadingSlash(newPath)}`,
              ),
            );
          }
        } else {
          await goto(`/files/${removeLeadingSlash(newPath)}`, {
            replaceState: true,
          });
        }
        closeModal();
      } catch (err) {
        error = err.response.data?.message;
      }
    },
  });

  $: runtimeInstanceId = $runtime.instanceId;
  $: existingDirectories = useDirectoryNamesInDirectory(
    runtimeInstanceId,
    folderName,
  );
  $: fileNamesInDirectory = useFileNamesInDirectory(
    runtimeInstanceId,
    folderName,
  );
</script>

<Dialog.Root
  open
  onOpenChange={(open) => {
    if (!open) {
      closeModal();
    }
  }}
  portal="#rill-portal"
>
  <Dialog.Content>
    <Dialog.Title>Rename</Dialog.Title>

    {#if $errors.newName?.[0]}
      <SubmissionError message={$errors.newName?.[0]} />
    {/if}
    <form
      id="rename-asset-form"
      class="flex flex-col gap-y-4"
      autocomplete="off"
      on:submit|preventDefault={submit}
      use:enhance
    >
      <Input
        bind:value={$superform.newName}
        claimFocusOnMount
        alwaysShowError
        id={isDir ? "folder-name" : "file-name"}
        label={isDir ? "Folder name" : "File name"}
        onEnter={submit}
      />
    </form>
    <Dialog.Footer class="gap-x-2">
      <Button large type="text" on:click={closeModal}>Cancel</Button>
      <Button large type="primary" submitForm form="rename-asset-form">
        Change Name
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
