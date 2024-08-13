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
        label: 'Learn',
        link: {
          type: 'doc',
          id: 'learn/index',

        },
        items: [
          
          {
            type: 'category',
            label: 'Rill Basics',
            description: 'Rill 100 series: Developer to Cloud & Developer features',
            items: [
                'learn/rill_learn_100/101_0',
                'learn/rill_learn_100/101_1',
                'learn/rill_learn_100/101_2',
                
                {
                  type: 'category',
                  label: 'Rill Developer Components',
                  items: [
                    'learn/rill_learn_100/components/102_0',
                    'learn/rill_learn_100/components/102_1',
                    'learn/rill_learn_100/components/102_2',
                    'learn/rill_learn_100/components/102_3',
                    'learn/rill_learn_100/components/102_4',
                    
                  ]
                },
                
                {
                  type: 'category',
                  label: 'Modifying the Metric-Views',
                  items: [
                    'learn/rill_learn_100/dashboard/103_0',
                    'learn/rill_learn_100/dashboard/103_1',
                    'learn/rill_learn_100/dashboard/103_2',
                    'learn/rill_learn_100/dashboard/103_3',
                    'learn/rill_learn_100/dashboard/103_0_u'
                  ]
                },
  
                    'learn/rill_learn_100/104_0',
              
  
                {
                  type: 'category',
                  label: 'Rill CLI',
                  items: [
                    'learn/rill_learn_100/CLI/105_0',
                    'learn/rill_learn_100/CLI/105_1',
                    'learn/rill_learn_100/CLI/105_2'
                  ]
                },
                {
                type: 'category',
                label: 'Deploy to Rill Cloud',
                items: [
                  'learn/rill_learn_100/deploy/106_0',
                  'learn/rill_learn_100/deploy/106_1'
                ]
              },
             'learn/rill_learn_100/107_0',
              ]
          },
          {
            type: 'category',
            label: 'Rill Advanced',
            description: 'Rill 200 series: Rill Cloud features, User Management, and more!',
            items: [
              'learn/rill_learn_200/201_0',
              'learn/rill_learn_200/201_1',
              'learn/rill_learn_200/202_0',
              
              'learn/rill_learn_200/203_0',
              {
                type: 'category',
                label: 'Back to Rill Developer',
                items: [
                  'learn/rill_learn_200/advanced_developer/204_0',
                  'learn/rill_learn_200/advanced_developer/204_1',
                  'learn/rill_learn_200/advanced_developer/204_2',
                ]
              },
              {
                type: 'category',
                label: 'Rill Cloud Features',
                items: [
                  'learn/rill_learn_200/cloud_components/205_0',
                  'learn/rill_learn_200/cloud_components/205_1',
                  'learn/rill_learn_200/cloud_components/205_2',
                  'learn/rill_learn_200/cloud_components/205_3',
                  'learn/rill_learn_200/cloud_components/205_4',
                ]
              },



              'learn/rill_learn_200/206_0',
              

              'learn/rill_learn_200/210_0',
            ]
          },
          {
            type: 'category',
            label: 'Rill Expert',
            description: 'Rill 300 series: Advanced use cases and beyond',
            items: [
              {
                type: 'category',
                label: 'Rill Custom Dashboards',
                items: [
                  'learn/rill_learn_300/custom_dashboards/301_0',
                  'learn/rill_learn_300/custom_dashboards/301_1',
                  'learn/rill_learn_300/custom_dashboards/301_2',
                  'learn/rill_learn_300/custom_dashboards/301_3',
                  'learn/rill_learn_300/custom_dashboards/301_4',       
                ]
              },
              {
                type: 'category',
                label: 'Incremental Models',
                items: [
                  'learn/rill_learn_300/incremental_models/302_0'
                  
                ]
              }
              ]
          },
          {
            type: 'category',
            label: 'Rill and ClickHouse',
            description: 'For our friends from ClickHouse, a revamped guide.',
            items: [
                'learn/rill_clickhouse/r_ch_0',
                'learn/rill_clickhouse/r_ch_1',
                'learn/rill_clickhouse/r_ch_2',
                'learn/rill_clickhouse/components/r_ch_3',
                'learn/rill_clickhouse/components/r_ch_4',
                'learn/rill_clickhouse/components/r_ch_5',
                {
                  type: 'category',
                  label: 'Deploy To Cloud',
                  link: {
                    type: 'doc',
                    id: 'learn/rill_clickhouse/r_ch_6',
          
                  },
                  items: [
                    'learn/rill_clickhouse/deploy/r_ch_7',
                    'learn/rill_clickhouse/deploy/r_ch_8',        
                  ]
                },

              ]
          },
        ],
      },
      {
        type: 'category',
        label: 'Guides',
        link: {
          type: 'doc',
          id: 'learn/guides/index',

        },
        items: [

          'learn/guides/use-case/all_in_one',
 

          {
            type: 'category',
            label: 'Cloud Storage',
            description: "See our guides specifically catered for cloud storage import",
            items: [
              'learn/guides/cloud-storage/S3-to-Rill',  
              'learn/guides/cloud-storage/GCS-to-Rill',
              'learn/guides/cloud-storage/ABS-to-Rill',
              
                          
            ]
          },
          {
            type: 'category',
            label: 'Data Warehouse',
            description: "See our guides specifically catered for cloud storage import",
            items: [
              'learn/guides/data-warehouse/BQ-to-Rill',  
                     
            ]
          },
          {
            type: 'category',
            label: 'OLAP',
            description: "In-depth into our different OLAP engines",
            items: [
              'learn/guides/OLAP/rill_clickhouse',           
            ]
          },
          {
            type: 'category',
            label: 'Concepts Explained',
            description: "Concepts further explained!",
            items: [
              'learn/guides/conceptual/avg_avg',  
              'learn/guides/conceptual/one_dashboard',      
            ]
          },


            'learn/guides/use-case/rill_on_rill',
            
        ],
  
    },
      
    ],
  
  
  
    refSidebar: [
      { type: "autogenerated", dirName: "reference" },
    ],
  };
  
  module.exports = sidebars;
  