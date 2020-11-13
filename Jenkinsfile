@Library("dst-shared") _

rpmBuild(
    channel: "loftsman-ci-alerts",
    slack_notify: ["FAILURE","FIXED"],
    product: "shasta-standard,shasta-premium",
    target_node: "ncn",
    fanout_params : ["sle15sp2"]
)
