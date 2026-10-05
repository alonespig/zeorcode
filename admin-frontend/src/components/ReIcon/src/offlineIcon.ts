// 这里存放本地图标，在 src/layout/index.vue 文件中加载，避免在首启动加载
import { getSvgInfo } from "@pureadmin/utils";
import { addIcon } from "@iconify/vue/dist/offline";

// https://icon-sets.iconify.design/ep/?keyword=ep
import EpHomeFilled from "~icons/ep/home-filled?raw";
import EpOdometer from "~icons/ep/odometer?raw";
import EpUser from "~icons/ep/user?raw";
import EpCpu from "~icons/ep/cpu?raw";
import EpBell from "~icons/ep/bell?raw";

// https://icon-sets.iconify.design/ri/?keyword=ri
import RiSearchLine from "~icons/ri/search-line?raw";
import RiInformationLine from "~icons/ri/information-line?raw";
import RiFileList3Line from "~icons/ri/file-list-3-line?raw";
import RiRobot2Line from "~icons/ri/robot-2-line?raw";

const icons = [
  // Element Plus Icon: https://github.com/element-plus/element-plus-icons
  ["ep/home-filled", EpHomeFilled],
  ["ep/odometer", EpOdometer],
  ["ep/user", EpUser],
  ["ep/cpu", EpCpu],
  ["ep/bell", EpBell],
  // Remix Icon: https://github.com/Remix-Design/RemixIcon
  ["ri/search-line", RiSearchLine],
  ["ri/information-line", RiInformationLine],
  ["ri/file-list-3-line", RiFileList3Line],
  ["ri/robot-2-line", RiRobot2Line]
];

// 本地菜单图标，后端在路由的 icon 中返回对应的图标字符串并且前端在此处使用 addIcon 添加即可渲染菜单图标
icons.forEach(([name, icon]) => {
  addIcon(name as string, getSvgInfo(icon as string));
});
