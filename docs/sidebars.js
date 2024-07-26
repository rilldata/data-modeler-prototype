/**
 * Creating a sidebar enables you to:
 * - create an ordered group of docs
 * - render a sidebar for each doc of that group
 * - provide next/previous navigation
 *
 * The sidebars can be generated from the filesystem, or explicitly defined here.
 *
 * Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docsSidebar: [
    { type: "autogenerated", dirName: "." },
  ],
  tutorialsSidebar: [
    {
      type: 'category',
      label: 'Tutorials',
      items: [
        {
          type: 'doc',
          id: 'tutorials/index'  // Main index page
          //className: 'hidden-page', // Custom className to hide from DocCardList
        },
        {
          type: 'category',
          label: 'Rill Basics',
          description: 'Basic tutorials description',
          items: [
              'tutorials/rill_learn_100/101_0',
              'tutorials/rill_learn_100/101_1',
              'tutorials/rill_learn_100/101_2',
              {
                type: 'category',
                label: 'Rill Developer Components',
                items: [
                  'tutorials/rill_learn_100/components/102_0',
                  'tutorials/rill_learn_100/components/102_1',
                  'tutorials/rill_learn_100/components/102_2',
                  'tutorials/rill_learn_100/components/102_3',
                  'tutorials/rill_learn_100/components/102_4',
                  
                ]
              },
              
              {
                type: 'category',
                label: 'Modifying the Metric-Views',
                items: [
                  {
                    type: 'category',
                    label: 'via YAML',
                    items: [
                      'tutorials/rill_learn_100/dashboard/103_0',
                      'tutorials/rill_learn_100/dashboard/103_1',
                      'tutorials/rill_learn_100/dashboard/103_2',
                       'tutorials/rill_learn_100/dashboard/103_3'
                    ]
                  },
                  {
                    type: 'category',
                    label: 'via UI',
                    items: [
                      'tutorials/rill_learn_100/dashboard/103_0_u'
                    ]
                  },

                ]
              },

                  'tutorials/rill_learn_100/104_0',
            

              {
                type: 'category',
                label: 'Rill CLI',
                items: [
                  'tutorials/rill_learn_100/CLI/105_0',
                  'tutorials/rill_learn_100/CLI/105_1',
                  'tutorials/rill_learn_100/CLI/105_2'
                ]
              },
              {
              type: 'category',
              label: 'Deploy to Rill Cloud',
              items: [
                {
                  type: 'category',
                  label: 'via CLI',
                  items: [
                     'tutorials/rill_learn_100/deploy/106_0',
                     'tutorials/rill_learn_100/deploy/106_1'
                  ]
                },
                {
                  type: 'category',
                  label: 'via UI',
                  items: [
                    'tutorials/rill_learn_100/deploy/106_0_u'
                  ]
                },

              ]
            },
            'tutorials/rill_learn_100/107_0',
           
            ]
        },
        {
          type: 'category',
          label: 'Rill Advanced',
          description: 'Advanced tutorials description',
          items: [
            'tutorials/rill_learn_200/201_0',
            'tutorials/rill_learn_200/202_0'
          ]
        },
        {
          type: 'category',
          label: 'Rill Expert',
          description: 'Expert tutorials description',
          items: [
              'tutorials/rill_learn_300/301_0',
              'tutorials/rill_learn_300/302_0'
            ]
        },
        {
          type: 'category',
          label: 'Rill and ClickHouse',
          description: 'For our friends from ClickHouse, a revamped guide.',
          items: [
              'tutorials/rill_clickhouse/r_ch_0',
              'tutorials/rill_clickhouse/r_ch_1',
              'tutorials/rill_clickhouse/r_ch_2',
            ]
        },
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        {
          type: 'doc',
          id: 'tutorials/guides'  
        },
        {
          type: 'category',
          label: 'Rill Basic Guides',
          description: 'Basic guide description',
          items: [
              'tutorials/guides/test'
          
          ],
        },
      ],

  },
    
  ],



  refSidebar: [
    { type: "autogenerated", dirName: "reference" },
  ],
};

module.exports = sidebars;
