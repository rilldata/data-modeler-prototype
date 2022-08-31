/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  // docsSidebar: [{ type: "autogenerated", dirName: "." }],

  docsSidebar: [
    { type: 'doc', id: 'overview/introduction', label: 'Introduction' },
    { type: 'category', label: 'Process Batch Data', items: ['get-started/processing-batch-data/aws-s3-bucket'],},
  ],
};

module.exports = sidebars;
